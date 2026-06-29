package services

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/mapper"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type ProductItemService interface {
	CreateProductItem(ctx context.Context, req *request.CreateProductItemRequest) (*response.ProductItemCreateResponse, error)
}

type productItemService struct {
	repo        repositories.ProductItemRepository
	batchRepo   batchRepo.BatchRepository
	variantRepo variantRepo.ProductVariantRepository
}

func NewProductItemService(
	repo repositories.ProductItemRepository,
	batchRepo batchRepo.BatchRepository,
	variantRepo variantRepo.ProductVariantRepository,
) ProductItemService {
	return &productItemService{
		repo:        repo,
		batchRepo:   batchRepo,
		variantRepo: variantRepo,
	}
}

// CreateProductItem tạo một product item mới.
// Các trường định danh (item_code, serial_number, verification_token) được
// tự động sinh tại tầng repository theo format đã quy định:
//
//   - item_code:          PTA-{YYMM}-{8 ký tự HEX viết hoa}   regex: ^PTA-\d{4}-[A-F0-9]{8}$
//   - serial_number:      SN{14 chữ số ngẫu nhiên}             regex: ^SN\d{14}$
//   - verification_token: MD5 hex 32 ký tự thường              regex: ^[a-f0-9]{32}$
func (s *productItemService) CreateProductItem(ctx context.Context, req *request.CreateProductItemRequest) (*response.ProductItemCreateResponse, error) {
	// FK check: batch_id phải tồn tại trong bảng batches (chỉ khi được truyền vào).
	if req.BatchID != uuid.Nil {
		batchExists, err := s.batchRepo.ExistsByID(ctx, req.BatchID)
		if err != nil {
			return nil, err
		}
		if !batchExists {
			return nil, apperror.NewNotFound("batch")
		}
	}

	// FK check: variant_id phải tồn tại trong bảng product_variants.
	variantExists, err := s.variantRepo.ExistsByID(ctx, req.VariantID)
	if err != nil {
		return nil, err
	}
	if !variantExists {
		return nil, apperror.NewNotFound("product_variant")
	}

	now := time.Now()
	newID := uuid.New()

	itemCode, err := generateItemCode(now)
	if err != nil {
		return nil, apperror.NewInternal("failed to generate item_code")
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, apperror.NewInternal("failed to generate serial_number")
	}

	verificationToken := generateVerificationToken(itemCode, serialNumber, newID, now)

	pi := mapper.CreateProductItemRequestToEntity(newID, itemCode, verificationToken, serialNumber, req)

	return s.repo.Create(ctx, pi)
}


// generateItemCode sinh item_code theo format PTA-{YYMM}-{8 ký tự HEX viết hoa}.
// Ví dụ: PTA-2501-686F493D
//
// Phần HEX 8 ký tự được tạo từ 4 byte ngẫu nhiên (crypto/rand) để đảm bảo
// tính duy nhất đủ cao trong phạm vi cùng tháng-năm.
func generateItemCode(now time.Time) (string, error) {
	yymm := now.Format("0601") // Go format: "06" = YY, "01" = MM

	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateItemCode: rand.Read: %w", err)
	}
	hexPart := fmt.Sprintf("%08X", binary.BigEndian.Uint32(b))

	return fmt.Sprintf("PTA-%s-%s", yymm, hexPart), nil
}

// generateSerialNumber sinh serial_number theo format SN{14 chữ số ngẫu nhiên}.
// Ví dụ: SN23650263663452
//
// Dùng crypto/rand để sinh 14 chữ số bằng cách lấy số ngẫu nhiên trong [0, 10^14).
func generateSerialNumber() (string, error) {
	// 10^14 = 100_000_000_000_000
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(14), nil)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("generateSerialNumber: rand.Int: %w", err)
	}
	// Pad với leading zero nếu cần để đảm bảo đúng 14 chữ số.
	return fmt.Sprintf("SN%014d", n), nil
}

// generateVerificationToken sinh verification_token là chuỗi MD5 hex 32 ký tự thường.
// Ví dụ: fe01edd8ce71cc2911b08118e76e1cd3
//
// Input của MD5 là tổ hợp: itemCode + serialNumber + uuid + timestamp (nano)
// để đảm bảo token duy nhất và không thể đoán ngược.
func generateVerificationToken(itemCode, serialNumber string, id uuid.UUID, now time.Time) string {
	raw := fmt.Sprintf("%s:%s:%s:%d", itemCode, serialNumber, id.String(), now.UnixNano())
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}
