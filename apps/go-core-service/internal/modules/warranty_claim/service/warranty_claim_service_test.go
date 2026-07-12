package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/service"
)

// Dummy implementations for mocks
type mockRepo struct {
    claim *entity.WarrantyClaim
}

func (m *mockRepo) Create(ctx context.Context, claim *entity.WarrantyClaim) error {
	m.claim = claim
	return nil
}
func (m *mockRepo) FindByProductItemIDAndStatusList(ctx context.Context, productItemID uuid.UUID, statuses []entity.WarrantyClaimStatus) (*entity.WarrantyClaim, error) {
	return m.claim, nil
}
func (m *mockRepo) FindByID(ctx context.Context, claimID uuid.UUID) (*entity.WarrantyClaim, error) {
	return m.claim, nil
}

type mockEvent struct{}
func (m *mockEvent) RecordEvent(ctx context.Context, productItemID uuid.UUID, eventType string, payload interface{}) error { return nil }

type mockAudit struct{}
func (m *mockAudit) LogAction(ctx context.Context, action string, userID uuid.UUID, resourceID uuid.UUID, details interface{}) error { return nil }

type mockNotification struct{}
func (m *mockNotification) NotifyStaff(ctx context.Context, claimID uuid.UUID, message string) error { return nil }

type mockOwnership struct{ isValid bool }
func (m *mockOwnership) VerifyOwnership(ctx context.Context, userID uuid.UUID, productItemID uuid.UUID) (bool, error) { return m.isValid, nil }

type mockProduct struct{ isValid bool }
func (m *mockProduct) CheckWarrantyValidity(ctx context.Context, productItemID uuid.UUID) (bool, error) { return m.isValid, nil }


func TestCreateWarrantyClaim_Success(t *testing.T) {
	svc := service.NewWarrantyClaimService(
		&mockRepo{claim: nil}, // no existing open claim
		&mockOwnership{isValid: true},
		&mockProduct{isValid: true},
		&mockEvent{},
		&mockAudit{},
		&mockNotification{},
	)
	
	resp, err := svc.CreateWarrantyClaim(context.Background(), uuid.New(), dto.CreateWarrantyClaimRequest{
		ProductID: uuid.New(),
		IssueTitle: "Broken Screen",
		IssueDescription: "Dropped it",
		ContactPhone: "0123456789",
	})
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("Expected response, got nil")
	}
	if resp.Status != entity.WarrantyClaimStatusOpen {
		t.Errorf("Expected status OPEN, got %v", resp.Status)
	}
}

func TestCreateWarrantyClaim_OwnershipFailed(t *testing.T) {
	svc := service.NewWarrantyClaimService(
		&mockRepo{claim: nil},
		&mockOwnership{isValid: false}, // Fails here
		&mockProduct{isValid: true},
		&mockEvent{},
		&mockAudit{},
		&mockNotification{},
	)
	
	_, err := svc.CreateWarrantyClaim(context.Background(), uuid.New(), dto.CreateWarrantyClaimRequest{
		ProductID: uuid.New(),
	})
	
	if err != service.ErrOwnershipValidation {
		t.Fatalf("Expected ErrOwnershipValidation, got %v", err)
	}
}

func TestCreateWarrantyClaim_WarrantyExpired(t *testing.T) {
	svc := service.NewWarrantyClaimService(
		&mockRepo{claim: nil},
		&mockOwnership{isValid: true}, // Passes
		&mockProduct{isValid: false}, // Fails here
		&mockEvent{},
		&mockAudit{},
		&mockNotification{},
	)
	
	_, err := svc.CreateWarrantyClaim(context.Background(), uuid.New(), dto.CreateWarrantyClaimRequest{
		ProductID: uuid.New(),
	})
	
	if err != service.ErrWarrantyExpired {
		t.Fatalf("Expected ErrWarrantyExpired, got %v", err)
	}
}

func TestCreateWarrantyClaim_AlreadyOpen(t *testing.T) {
	svc := service.NewWarrantyClaimService(
		&mockRepo{claim: &entity.WarrantyClaim{ID: uuid.New(), CreatedAt: time.Now()}}, // Exist OPEN claim
		&mockOwnership{isValid: true}, 
		&mockProduct{isValid: true},
		&mockEvent{},
		&mockAudit{},
		&mockNotification{},
	)
	
	_, err := svc.CreateWarrantyClaim(context.Background(), uuid.New(), dto.CreateWarrantyClaimRequest{
		ProductID: uuid.New(),
	})
	
	if err != service.ErrClaimAlreadyOpen {
		t.Fatalf("Expected ErrClaimAlreadyOpen, got %v", err)
	}
}
