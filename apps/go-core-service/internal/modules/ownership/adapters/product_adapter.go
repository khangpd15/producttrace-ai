package adapters

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
)

type DummyProductAdapter struct{}

func NewDummyProductAdapter() service.IProductService {
	return &DummyProductAdapter{}
}

func (a *DummyProductAdapter) FindProductByQR(ctx context.Context, qrCode string) (uuid.UUID, error) {
	log.Printf("[ProductAdapter] Faking FindProductByQR for QR: %s\n", qrCode)
	// Return a dummy UUID
	return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
}

func (a *DummyProductAdapter) ValidateProductOwnershipStatus(ctx context.Context, productID uuid.UUID) error {
	log.Printf("[ProductAdapter] Faking ValidateProductOwnershipStatus for Product: %s\n", productID)
	return nil
}

func (a *DummyProductAdapter) UpdateOwnershipStatus(ctx context.Context, productID uuid.UUID, status string) error {
	log.Printf("[ProductAdapter] Faking UpdateOwnershipStatus for Product: %s to Status: %s\n", productID, status)
	return nil
}

func (a *DummyProductAdapter) GetProductItemDetail(ctx context.Context, productItemID uuid.UUID) (string, string, error) {
	log.Printf("[ProductAdapter] Faking GetProductItemDetail for ProductItem: %s\n", productItemID)
	return "Sản phẩm Demo " + productItemID.String()[:8], "SKU-DEMO-12345", nil
}
