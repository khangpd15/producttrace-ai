package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/adapters"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

var (
	ErrWarrantyNotFound    = errors.New("warranty record not found")
	ErrInvalidStatus       = errors.New("invalid status transition")
	ErrProductItemNotFound = errors.New("product item not found")
)

type WarrantyService interface {
	RequestActivation(ctx context.Context, userID uuid.UUID, req dto.RequestWarrantyActivationRequest) (*dto.WarrantyResponse, error)
	ConfirmActivation(ctx context.Context, actorID uuid.UUID, id uuid.UUID, req dto.ConfirmWarrantyRequest) (*dto.WarrantyResponse, error)
	GetWarrantyByID(ctx context.Context, id uuid.UUID) (*dto.WarrantyResponse, error)
	ListWarranties(ctx context.Context, search string, status string, ownerID *uuid.UUID, page, limit int) (*dto.WarrantyListResponse, error)
	UpdateWarranty(ctx context.Context, actorID uuid.UUID, id uuid.UUID, req dto.UpdateWarrantyRequest) (*dto.WarrantyResponse, error)
	DeleteWarranty(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error
	GetStats(ctx context.Context) (map[string]int64, error)
}

type warrantyService struct {
	repo             repository.WarrantyRepository
	ownershipPort    adapters.OwnershipPort
	eventPort        adapters.EventPort
	auditLogPort     adapters.AuditLogPort
	notificationPort adapters.NotificationPort
	db               *gorm.DB
}

func NewWarrantyService(
	repo repository.WarrantyRepository,
	ownershipPort adapters.OwnershipPort,
	eventPort adapters.EventPort,
	auditLogPort adapters.AuditLogPort,
	notificationPort adapters.NotificationPort,
	db *gorm.DB,
) WarrantyService {
	return &warrantyService{
		repo:             repo,
		ownershipPort:    ownershipPort,
		eventPort:        eventPort,
		auditLogPort:     auditLogPort,
		notificationPort: notificationPort,
		db:               db,
	}
}

func (s *warrantyService) RequestActivation(ctx context.Context, userID uuid.UUID, req dto.RequestWarrantyActivationRequest) (*dto.WarrantyResponse, error) {
	// 1. Verify Ownership
	isOwner, err := s.ownershipPort.VerifyOwnership(ctx, userID, req.ProductItemID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, apperror.Wrap(errors.New("ownership verification failed"), apperror.NewForbidden("Sản phẩm chưa có chủ sở hữu hoặc bạn không phải là chủ sở hữu của sản phẩm này."))
	}

	// 2. Check if a warranty request already exists for this item
	existing, err := s.repo.FindByProductItemID(ctx, req.ProductItemID)
	if err != nil {
		return nil, err
	}
	if existing != nil && (existing.Status == entity.WarrantyStatusActive || existing.Status == entity.WarrantyStatusInactive) {
		return nil, apperror.NewConflict("Yêu cầu bảo hành cho sản phẩm này đã được gửi hoặc đang hoạt động.")
	}

	// 3. Create inactive warranty record
	w := &entity.Warranty{
		ID:            uuid.New(),
		ProductItemID: req.ProductItemID,
		OwnerID:       &userID,
		Status:        entity.WarrantyStatusInactive,
		InvoiceNumber: req.InvoiceNumber,
		InvoiceURL:    req.InvoiceURL,
		Note:          req.Note,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}

	// 4. Record Event (Async/Sync)
	_ = s.eventPort.RecordEvent(ctx, req.ProductItemID, "WARRANTY_ACTIVATION_REQUESTED", map[string]string{
		"warranty_id":    w.ID.String(),
		"invoice_number": w.InvoiceNumber,
	})

	// 5. Audit Log
	_ = s.auditLogPort.LogAction(ctx, "REQUEST_WARRANTY_ACTIVATION", userID, w.ID, "Customer requested warranty activation")

	// 6. Notify Staff
	_ = s.notificationPort.NotifyStaff(ctx, w.ID, fmt.Sprintf("Yêu cầu kích hoạt bảo hành mới cho sản phẩm %s", w.ProductItemID))

	return s.mapToResponse(w), nil
}

func (s *warrantyService) ConfirmActivation(ctx context.Context, actorID uuid.UUID, id uuid.UUID, req dto.ConfirmWarrantyRequest) (*dto.WarrantyResponse, error) {
	// Start a transaction for state mutation and product item status updates
	var w *entity.Warranty
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		w, err = s.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if w == nil {
			return apperror.NewNotFound("Warranty not found")
		}

		if w.Status != entity.WarrantyStatusInactive {
			return apperror.Wrap(ErrInvalidStatus, apperror.NewBadRequest("Yêu cầu bảo hành không còn ở trạng thái chờ kích hoạt."))
		}

		if req.Decision == "APPROVE" {
			w.Status = entity.WarrantyStatusActive
			w.WarrantyCode = fmt.Sprintf("WRY-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
			
			duration := 12
			if req.DurationMonths > 0 {
				duration = req.DurationMonths
			}
			w.DurationMonths = duration
			
			now := time.Now()
			w.StartDate = &now
			end := now.AddDate(0, duration, 0)
			w.EndDate = &end
			w.ActivatedAt = &now
			w.PolicyName = req.PolicyName
			if w.PolicyName == "" {
				w.PolicyName = "Tiêu chuẩn"
			}
			w.PolicyDescription = req.PolicyDescription
			if w.PolicyDescription == "" {
				w.PolicyDescription = fmt.Sprintf("Chính sách bảo hành tiêu chuẩn %d tháng", duration)
			}
			if req.Note != "" {
				w.Note = req.Note
			}

			// Update product item status to WARRANTY_ACTIVE
			err = tx.Table("product_items").Where("id = ?", w.ProductItemID).Update("status", "WARRANTY_ACTIVE").Error
			if err != nil {
				return apperror.WrapDBError(err, "product_items")
			}

			// Record Event
			_ = s.eventPort.RecordEvent(ctx, w.ProductItemID, "WARRANTY_ACTIVATED", map[string]string{
				"warranty_id":   w.ID.String(),
				"warranty_code": w.WarrantyCode,
				"end_date":      w.EndDate.Format(time.RFC3339),
			})

			// Audit Log
			_ = s.auditLogPort.LogAction(ctx, "APPROVE_WARRANTY_ACTIVATION", actorID, w.ID, "Approved warranty activation")

			// Notify Customer
			if w.OwnerID != nil {
				_ = s.notificationPort.NotifyCustomer(ctx, *w.OwnerID, "Bảo hành đã được kích hoạt", fmt.Sprintf("Sản phẩm %s đã được kích hoạt bảo hành thành công. Mã bảo hành: %s", w.ProductItemID, w.WarrantyCode))
			}

		} else if req.Decision == "REJECT" {
			w.Status = entity.WarrantyStatusRejected
			if req.Note != "" {
				w.Note = req.Note
			}

			// Record Event
			_ = s.eventPort.RecordEvent(ctx, w.ProductItemID, "WARRANTY_ACTIVATION_REJECTED", map[string]string{
				"warranty_id": w.ID.String(),
				"reason":      w.Note,
			})

			// Audit Log
			_ = s.auditLogPort.LogAction(ctx, "REJECT_WARRANTY_ACTIVATION", actorID, w.ID, "Rejected warranty activation")

			// Notify Customer
			if w.OwnerID != nil {
				_ = s.notificationPort.NotifyCustomer(ctx, *w.OwnerID, "Yêu cầu kích hoạt bảo hành bị từ chối", fmt.Sprintf("Yêu cầu kích hoạt bảo hành cho sản phẩm %s đã bị từ chối. Lý do: %s", w.ProductItemID, w.Note))
			}
		}

		w.UpdatedAt = time.Now()
		return tx.Save(w).Error
	})

	if err != nil {
		return nil, err
	}

	return s.mapToResponse(w), nil
}

func (s *warrantyService) GetWarrantyByID(ctx context.Context, id uuid.UUID) (*dto.WarrantyResponse, error) {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, apperror.NewNotFound("Warranty not found")
	}
	
	// Fetch item and owner names if not scanned (GORM scan does it, but double check)
	if w.ItemCode == "" {
		var item struct {
			ItemCode     string
			SerialNumber string
		}
		err = s.db.WithContext(ctx).Table("product_items").Select("item_code, serial_number").Where("id = ?", w.ProductItemID).Scan(&item).Error
		if err == nil {
			w.ItemCode = item.ItemCode
			w.SerialNumber = item.SerialNumber
		}
	}

	if w.OwnerID != nil && w.OwnerName == "" {
		var user struct {
			FullName string
			Email    string
		}
		err = s.db.WithContext(ctx).Table("users").Select("full_name, email").Where("id = ?", *w.OwnerID).Scan(&user).Error
		if err == nil {
			w.OwnerName = user.FullName
			w.OwnerEmail = user.Email
		}
	}

	return s.mapToResponse(w), nil
}

func (s *warrantyService) ListWarranties(ctx context.Context, search string, status string, ownerID *uuid.UUID, page, limit int) (*dto.WarrantyListResponse, error) {
	// Auto expire any outdated active warranties first
	_, _ = s.repo.UpdateExpiredStatus(ctx)

	list, total, err := s.repo.FindAll(ctx, search, status, ownerID, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]*dto.WarrantyResponse, len(list))
	for i, w := range list {
		items[i] = s.mapToResponse(w)
	}

	return &dto.WarrantyListResponse{
		Items: items,
		Total: total,
	}, nil
}

func (s *warrantyService) UpdateWarranty(ctx context.Context, actorID uuid.UUID, id uuid.UUID, req dto.UpdateWarrantyRequest) (*dto.WarrantyResponse, error) {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, apperror.NewNotFound("Warranty not found")
	}

	oldW := *w

	if req.PolicyName != "" {
		w.PolicyName = req.PolicyName
	}
	if req.PolicyDescription != "" {
		w.PolicyDescription = req.PolicyDescription
	}
	if req.DurationMonths > 0 {
		w.DurationMonths = req.DurationMonths
	}
	if req.StartDate != nil {
		w.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		w.EndDate = req.EndDate
	}
	if req.Status != "" {
		w.Status = entity.WarrantyStatus(req.Status)
	}
	if req.Note != "" {
		w.Note = req.Note
	}
	w.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.auditLogPort.LogAction(ctx, "UPDATE_WARRANTY", actorID, w.ID, map[string]interface{}{
		"old": oldW,
		"new": w,
	})

	return s.mapToResponse(w), nil
}

func (s *warrantyService) DeleteWarranty(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	w, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if w == nil {
		return apperror.NewNotFound("Warranty not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit Log
	_ = s.auditLogPort.LogAction(ctx, "DELETE_WARRANTY", actorID, id, fmt.Sprintf("Deleted warranty with code %s", w.WarrantyCode))

	return nil
}

func (s *warrantyService) GetStats(ctx context.Context) (map[string]int64, error) {
	return s.repo.GetStats(ctx)
}

func (s *warrantyService) mapToResponse(w *entity.Warranty) *dto.WarrantyResponse {
	return &dto.WarrantyResponse{
		ID:                w.ID,
		ProductItemID:     w.ProductItemID,
		OwnerID:           w.OwnerID,
		WarrantyCode:      w.WarrantyCode,
		PolicyName:        w.PolicyName,
		PolicyDescription: w.PolicyDescription,
		DurationMonths:    w.DurationMonths,
		Status:            w.Status,
		StartDate:         w.StartDate,
		EndDate:           w.EndDate,
		ActivatedAt:       w.ActivatedAt,
		InvoiceNumber:     w.InvoiceNumber,
		InvoiceURL:        w.InvoiceURL,
		Note:              w.Note,
		ItemCode:          w.ItemCode,
		SerialNumber:      w.SerialNumber,
		OwnerName:         w.OwnerName,
		OwnerEmail:        w.OwnerEmail,
		CreatedAt:         w.CreatedAt,
		UpdatedAt:         w.UpdatedAt,
	}
}
