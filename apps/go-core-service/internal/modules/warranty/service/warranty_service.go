package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	auditlog "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
	"gorm.io/gorm"
)

var (
	ErrWarrantyExists   = errors.New("warranty already exists for this product")
	ErrWarrantyNotFound = errors.New("warranty not found")
	ErrNotOwned         = errors.New("product is not registered for ownership or ownership is inactive")
	ErrOver30Days       = errors.New("warranty registration request exceeds 30 days from ownership date")
)

type WarrantyService interface {
	ActivateWarranty(ctx context.Context, req dto.CreateWarrantyRequest) (*dto.WarrantyResponse, error)
	RequestWarranty(ctx context.Context, req dto.CustomerRequestWarrantyRequest) (*dto.WarrantyResponse, error)
	ApproveWarranty(ctx context.Context, id uuid.UUID, req dto.ApproveWarrantyRequest) (*dto.WarrantyResponse, error)
	RejectWarranty(ctx context.Context, id uuid.UUID, req dto.RejectWarrantyRequest) (*dto.WarrantyResponse, error)
	ListWarranties(ctx context.Context) ([]dto.WarrantyResponse, error)
	GetWarrantyByCode(ctx context.Context, code string) (*dto.WarrantyResponse, error)
	GetWarrantyByID(ctx context.Context, id uuid.UUID) (*dto.WarrantyResponse, error)
	GetWarrantyByProductItemID(ctx context.Context, productItemID uuid.UUID) (*dto.WarrantyResponse, error)
	VoidWarranty(ctx context.Context, id uuid.UUID, reason string) (*dto.WarrantyResponse, error)
	ListMyWarranties(ctx context.Context, userID uuid.UUID) ([]dto.WarrantyResponse, error)
}

type warrantyService struct {
	repo     repository.WarrantyRepository
	pub      *publisher.Publisher
	auditLog auditlog.AuditLogService
}

func NewWarrantyService(repo repository.WarrantyRepository, pub *publisher.Publisher, auditLog auditlog.AuditLogService) WarrantyService {
	return &warrantyService{repo: repo, pub: pub, auditLog: auditLog}
}

func getActorUUID(ctx context.Context) *uuid.UUID {
	actorID := utils.GetActorID(ctx)
	if actorID == "" {
		return nil
	}
	u, err := uuid.Parse(actorID)
	if err != nil {
		return nil
	}
	return &u
}

func (s *warrantyService) ActivateWarranty(ctx context.Context, req dto.CreateWarrantyRequest) (*dto.WarrantyResponse, error) {
	// 1. Verify active ownership
	ownershipInfo, err := s.repo.GetActiveOwnership(ctx, req.ItemCode, req.SerialNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotOwned
		}
		return nil, err
	}

	// 2. Check if a warranty already exists for this serial number
	existing, err := s.repo.FindBySerialNumber(ctx, req.SerialNumber)
	if err == nil && len(existing) > 0 {
		return nil, ErrWarrantyExists
	}

	// 3. Check if warranty code already exists
	_, err = s.repo.FindByWarrantyCode(ctx, req.WarrantyCode)
	if err == nil {
		return nil, ErrWarrantyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 4. Parse dates
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", req.StartDate)
	} else {
		startDate = time.Now()
	}

	if req.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", req.EndDate)
	} else {
		endDate = startDate.AddDate(0, req.DurationMonths, 0)
	}

	status := entity.WarrantyStatusActive
	if req.Status != "" {
		status = entity.WarrantyStatus(req.Status)
	}

	ownerName := req.OwnerName
	if ownerName == "" {
		ownerName = ownershipInfo.OwnerName
	}
	ownerEmail := req.OwnerEmail
	if ownerEmail == "" {
		ownerEmail = ownershipInfo.OwnerEmail
	}

	// 5. Create entity with ProductItemID and OwnerID
	warranty := &entity.Warranty{
		ID:                uuid.New(),
		ProductItemID:     ownershipInfo.ProductItemID,
		OwnerID:           &ownershipInfo.OwnerID,
		ItemCode:          req.ItemCode,
		ItemName:          req.ItemName,
		SerialNumber:      req.SerialNumber,
		OwnerName:         ownerName,
		OwnerEmail:        ownerEmail,
		WarrantyCode:      req.WarrantyCode,
		PolicyName:        req.PolicyName,
		PolicyDescription: req.PolicyDescription,
		DurationMonths:    req.DurationMonths,
		Status:            status,
		StartDate:         startDate,
		EndDate:           endDate,
		InvoiceNumber:     req.InvoiceNumber,
		Note:              req.Note,
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(warranty).Error; err != nil {
			return err
		}

		if status == entity.WarrantyStatusActive {
			// Update product_items status -> 'WARRANTY_ACTIVE'
			if err := tx.Table("product_items").
				Where("id = ? AND is_deleted = false", warranty.ProductItemID).
				Update("status", "WARRANTY_ACTIVE").Error; err != nil {
				return err
			}

			// Insert Event to database
			eventObj := map[string]interface{}{
				"id":              uuid.New(),
				"product_item_id": warranty.ProductItemID,
				"event_type":      "WARRANTY_ACTIVE",
				"title":           "Kích hoạt bảo hành",
				"description":     "Bảo hành sản phẩm được tạo trực tiếp và kích hoạt thành công",
				"created_at":      time.Now(),
			}
			if err := tx.Table("events").Create(&eventObj).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish to RabbitMQ after commit
	if status == entity.WarrantyStatusActive && s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     rabbitmq.WarrantyCreatedRK,
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"warranty_id":   warranty.ID.String(),
				"item_code":     warranty.ItemCode,
				"serial_number": warranty.SerialNumber,
				"owner_name":    warranty.OwnerName,
				"owner_email":   warranty.OwnerEmail,
				"policy_name":   warranty.PolicyName,
				"status":        "ACTIVE",
				"start_date":    warranty.StartDate,
				"end_date":      warranty.EndDate,
			},
		}
		_ = s.pub.Publish(event)
	}

	actorUUID := getActorUUID(ctx)
	if s.auditLog != nil {
		_ = s.auditLog.LogCreate(ctx, actorUUID, "Warranty", warranty.ID, warranty)
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) RequestWarranty(ctx context.Context, req dto.CustomerRequestWarrantyRequest) (*dto.WarrantyResponse, error) {
	// 1. Verify active ownership
	ownershipInfo, err := s.repo.GetActiveOwnership(ctx, req.ItemCode, req.SerialNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotOwned
		}
		return nil, err
	}

	// 2. Check if a warranty already exists for this serial number
	existing, err := s.repo.FindBySerialNumber(ctx, req.SerialNumber)
	if err == nil && len(existing) > 0 {
		return nil, ErrWarrantyExists
	}

	// 3. Create Pending Warranty
	warranty := &entity.Warranty{
		ID:            uuid.New(),
		ProductItemID: ownershipInfo.ProductItemID,
		OwnerID:       &ownershipInfo.OwnerID,
		ItemCode:      req.ItemCode,
		SerialNumber:  req.SerialNumber,
		OwnerName:     req.OwnerName,
		OwnerEmail:    req.OwnerEmail,
		WarrantyCode:  "WAR-" + uuid.New().String()[:8], // Generate a random code or wait for admin
		Status:        entity.WarrantyStatusPending,
		Note:          req.Note,
	}

	if err := s.repo.Create(ctx, warranty); err != nil {
		return nil, err
	}

	actorUUID := getActorUUID(ctx)
	if s.auditLog != nil {
		_ = s.auditLog.LogCreate(ctx, actorUUID, "Warranty", warranty.ID, warranty)
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) ApproveWarranty(ctx context.Context, id uuid.UUID, req dto.ApproveWarrantyRequest) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}

	if warranty.Status != entity.WarrantyStatusPending {
		return nil, errors.New("only pending warranties can be approved")
	}

	oldWarranty := *warranty

	ownedAt, err := s.repo.GetOwnershipDate(ctx, warranty.ItemCode, warranty.SerialNumber)
	if err != nil {
		return nil, errors.New("could not verify ownership")
	}

	if time.Since(*ownedAt) > 30*24*time.Hour {
		return nil, ErrOver30Days
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, req.DurationMonths, 0)

	warranty.Status = entity.WarrantyStatusActive
	warranty.PolicyName = req.PolicyName
	warranty.DurationMonths = req.DurationMonths
	warranty.StartDate = startDate
	warranty.EndDate = endDate

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Save(warranty).Error; err != nil {
			return err
		}

		// Update product_items status -> 'WARRANTY_ACTIVE'
		if err := tx.Table("product_items").
			Where("id = ? AND is_deleted = false", warranty.ProductItemID).
			Update("status", "WARRANTY_ACTIVE").Error; err != nil {
			return err
		}

		// Insert Event to database
		eventObj := map[string]interface{}{
			"id":              uuid.New(),
			"product_item_id": warranty.ProductItemID,
			"event_type":      "WARRANTY_ACTIVE",
			"title":           "Kích hoạt bảo hành",
			"description":     "Bảo hành sản phẩm được kích hoạt thành công",
			"created_at":      time.Now(),
		}
		if err := tx.Table("events").Create(&eventObj).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish to RabbitMQ after commit
	if s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     rabbitmq.WarrantyCreatedRK,
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"warranty_id":   warranty.ID.String(),
				"item_code":     warranty.ItemCode,
				"serial_number": warranty.SerialNumber,
				"owner_name":    warranty.OwnerName,
				"owner_email":   warranty.OwnerEmail,
				"policy_name":   warranty.PolicyName,
				"status":        "ACTIVE",
				"start_date":    warranty.StartDate,
				"end_date":      warranty.EndDate,
			},
		}
		_ = s.pub.Publish(event)
	}

	actorUUID := getActorUUID(ctx)
	if s.auditLog != nil {
		_ = s.auditLog.LogUpdate(ctx, actorUUID, "Warranty", warranty.ID, oldWarranty, warranty)
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) RejectWarranty(ctx context.Context, id uuid.UUID, req dto.RejectWarrantyRequest) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}

	if warranty.Status != entity.WarrantyStatusPending {
		return nil, errors.New("only pending warranties can be rejected")
	}

	oldWarranty := *warranty

	warranty.Status = entity.WarrantyStatusRejected
	warranty.Note = req.Reason

	if err := s.repo.Update(ctx, warranty); err != nil {
		return nil, err
	}

	actorUUID := getActorUUID(ctx)
	if s.auditLog != nil {
		_ = s.auditLog.LogUpdate(ctx, actorUUID, "Warranty", warranty.ID, oldWarranty, warranty)
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) ListWarranties(ctx context.Context) ([]dto.WarrantyResponse, error) {
	warranties, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return dto.FromWarrantyEntities(warranties), nil
}

func (s *warrantyService) GetWarrantyByCode(ctx context.Context, code string) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByWarrantyCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}
	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) GetWarrantyByID(ctx context.Context, id uuid.UUID) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}
	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) GetWarrantyByProductItemID(ctx context.Context, productItemID uuid.UUID) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByProductItemID(ctx, productItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}
	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) VoidWarranty(ctx context.Context, id uuid.UUID, reason string) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}

	oldWarranty := *warranty
	warranty.Status = entity.WarrantyStatusCancelled
	if reason != "" {
		if warranty.Note != "" {
			warranty.Note = warranty.Note + " | Lý do hủy: " + reason
		} else {
			warranty.Note = "Lý do hủy: " + reason
		}
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Save(warranty).Error; err != nil {
			return err
		}

		// Update product_items status back to 'REGISTERED' since warranty is voided
		if err := tx.Table("product_items").
			Where("id = ? AND is_deleted = false", warranty.ProductItemID).
			Update("status", "REGISTERED").Error; err != nil {
			return err
		}

		// Insert Event to database of type WARRANTY_RESOLVED
		eventObj := map[string]interface{}{
			"id":              uuid.New(),
			"product_item_id": warranty.ProductItemID,
			"event_type":      "WARRANTY_RESOLVED",
			"title":           "Hủy bảo hành",
			"description":     "Bảo hành sản phẩm bị hủy / vô hiệu hóa. Lý do: " + reason,
			"created_at":      time.Now(),
		}
		if err := tx.Table("events").Create(&eventObj).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	actorUUID := getActorUUID(ctx)
	if s.auditLog != nil {
		_ = s.auditLog.LogUpdate(ctx, actorUUID, "Warranty", warranty.ID, oldWarranty, warranty)
	}

	// Publish to RabbitMQ after commit
	if s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     "warranty.voided",
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"warranty_id":   warranty.ID.String(),
				"item_code":     warranty.ItemCode,
				"serial_number": warranty.SerialNumber,
				"status":        "CANCELLED",
				"voided_at":     time.Now().UTC(),
				"reason":        reason,
			},
		}
		_ = s.pub.Publish(event)
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) ListMyWarranties(ctx context.Context, userID uuid.UUID) ([]dto.WarrantyResponse, error) {
	warranties, err := s.repo.FindMyWarranties(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.FromWarrantyEntities(warranties), nil
}
