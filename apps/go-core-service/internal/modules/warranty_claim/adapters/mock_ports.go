package adapters

import (
	"context"
	"log"

	"github.com/google/uuid"
)

// Mock representations for missing modules

type MockEventPort struct{}

func (m *MockEventPort) RecordEvent(ctx context.Context, productItemID uuid.UUID, eventType string, payload interface{}) error {
	log.Printf("[MOCK EVENT] Product: %s | Event: %s | Payload: %+v\n", productItemID, eventType, payload)
	return nil
}

type MockAuditLogPort struct{}

func (m *MockAuditLogPort) LogAction(ctx context.Context, action string, userID uuid.UUID, resourceID uuid.UUID, details interface{}) error {
	log.Printf("[MOCK AUDIT] User: %s | Action: %s | Resource: %s | Details: %+v\n", userID, action, resourceID, details)
	return nil
}

type MockNotificationPort struct{}

func (m *MockNotificationPort) NotifyStaff(ctx context.Context, claimID uuid.UUID, message string) error {
	log.Printf("[MOCK NOTIFICATION] ClaimID: %s | Msg: %s\n", claimID, message)
	return nil
}

type MockOwnershipPort struct{}

func (m *MockOwnershipPort) VerifyOwnership(ctx context.Context, userID uuid.UUID, productItemID uuid.UUID) (bool, error) {
	log.Printf("[MOCK OWNERSHIP] User: %s | Product: %s (Auto-approved for now)\n", userID, productItemID)
	return true, nil // Always allow for mock
}

type MockProductItemPort struct{}

func (m *MockProductItemPort) CheckWarrantyValidity(ctx context.Context, productItemID uuid.UUID) (bool, error) {
	log.Printf("[MOCK PRODUCT] Check Warranty for product: %s (Auto-approved for now)\n", productItemID)
	return true, nil // Always true for mock
}
