package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/adapters"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrOwnershipValidation = errors.New("ownership validation failed or access denied")
	ErrWarrantyExpired     = errors.New("warranty expired or not valid")
	ErrClaimAlreadyOpen    = errors.New("a warranty claim is already open for this product")
)

type WarrantyClaimService interface {
	CreateWarrantyClaim(ctx context.Context, userID uuid.UUID, req dto.CreateWarrantyClaimRequest) (*dto.WarrantyClaimResponse, error)
}

type warrantyClaimService struct {
	repo             repository.WarrantyClaimRepository
	ownershipPort    adapters.OwnershipPort
	productPort      adapters.ProductItemPort
	eventPort        adapters.EventPort
	auditLogPort     adapters.AuditLogPort
	notificationPort adapters.NotificationPort
}

func NewWarrantyClaimService(
	repo repository.WarrantyClaimRepository,
	ownershipPort adapters.OwnershipPort,
	productPort adapters.ProductItemPort,
	eventPort adapters.EventPort,
	auditLogPort adapters.AuditLogPort,
	notificationPort adapters.NotificationPort,
) WarrantyClaimService {
	return &warrantyClaimService{
		repo:             repo,
		ownershipPort:    ownershipPort,
		productPort:      productPort,
		eventPort:        eventPort,
		auditLogPort:     auditLogPort,
		notificationPort: notificationPort,
	}
}

func (s *warrantyClaimService) CreateWarrantyClaim(ctx context.Context, userID uuid.UUID, req dto.CreateWarrantyClaimRequest) (*dto.WarrantyClaimResponse, error) {
	// 1. Verify Ownership
	ownerID, err := s.ownershipPort.GetActiveOwner(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Wrap(ErrOwnershipValidation, apperror.NewBadRequest("Sản phẩm chưa có chủ sở hữu. Vui lòng đăng ký sở hữu trước."))
		}
		return nil, err
	}
	if ownerID != userID {
		return nil, apperror.Wrap(ErrOwnershipValidation, apperror.NewForbidden("Bạn không phải chủ sở hữu hiện tại."))
	}

	// 2. Validate Warranty validity
	isValid, err := s.productPort.CheckWarrantyValidity(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, apperror.Wrap(ErrWarrantyExpired, apperror.NewBadRequest("Sản phẩm đã hết hạn bảo hành hoặc không hợp lệ."))
	}

	// 3. Check for open claims
	openClaim, err := s.repo.FindByProductItemIDAndStatusList(ctx, req.ProductID, []entity.WarrantyClaimStatus{
		entity.WarrantyClaimStatusOpen,
		entity.WarrantyClaimStatusProcessing,
	})
	if err != nil {
		return nil, err
	}
	if openClaim != nil {
		return nil, apperror.Wrap(ErrClaimAlreadyOpen, apperror.NewConflict("Sản phẩm đang có yêu cầu bảo hành đang được xử lý."))
	}

	// 4. Create Claim Entity
	var attachmentsBytes []byte
	if len(req.Attachments) > 0 {
		attachmentsBytes, _ = json.Marshal(req.Attachments)
	}

	claimNumber := fmt.Sprintf("CLM-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	
	var serviceCenterID *uuid.UUID
	if req.PreferredServiceCenterID != "" {
		parsedCenter, err := uuid.Parse(req.PreferredServiceCenterID)
		if err == nil {
			serviceCenterID = &parsedCenter
		}
	}

	claim := &entity.WarrantyClaim{
		ID:                       uuid.New(),
		ClaimNumber:              claimNumber,
		ProductItemID:            req.ProductID,
		IssueTitle:               req.IssueTitle,
		IssueDescription:         req.IssueDescription,
		ContactPhone:             req.ContactPhone,
		PreferredServiceCenterID: serviceCenterID,
		Status:                   entity.WarrantyClaimStatusOpen,
		CreatedBy:                userID,
	}
	
	if len(attachmentsBytes) > 0 {
		claim.AttachmentsJSON = datatypes.JSON(attachmentsBytes)
	}
	
	if req.ContactEmail != "" {
		claim.ContactEmail = &req.ContactEmail
	}

	// 5. Save Claim
	if err := s.repo.Create(ctx, claim); err != nil {
		return nil, err
	}

	// 6. Record Event (Async or Sync)
	_ = s.eventPort.RecordEvent(ctx, claim.ProductItemID, "WARRANTY_CLAIM_CREATED", map[string]string{
		"claim_id": claim.ID.String(),
		"claim_number": claim.ClaimNumber,
	})

	// 7. Audit Log
	_ = s.auditLogPort.LogAction(ctx, "CREATE_WARRANTY_CLAIM", userID, claim.ID, "Created new warranty claim")

	// 8. Send Notification
	_ = s.notificationPort.NotifyStaff(ctx, claim.ID, fmt.Sprintf("New warranty claim %s created for product %s", claim.ClaimNumber, claim.ProductItemID))

	return &dto.WarrantyClaimResponse{
		ID:                       claim.ID,
		ClaimNumber:              claim.ClaimNumber,
		ProductItemID:            claim.ProductItemID,
		IssueTitle:               claim.IssueTitle,
		IssueDescription:         claim.IssueDescription,
		ContactPhone:             claim.ContactPhone,
		ContactEmail:             claim.ContactEmail,
		PreferredServiceCenterID: claim.PreferredServiceCenterID,
		Attachments:              req.Attachments,
		Status:                   claim.Status,
		CreatedAt:                claim.CreatedAt,
		UpdatedAt:                claim.UpdatedAt,
	}, nil
}
