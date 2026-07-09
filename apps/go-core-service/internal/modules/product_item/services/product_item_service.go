package services

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/consumer"
	batchDTO "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/mapper"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	amqp "github.com/rabbitmq/amqp091-go"
)

// batchEventEnvelope là struct dùng để unwrap message từ publisher.
// Publisher gửi theo NestJS format:
//
//	{
//	  "pattern": "batch.created",
//	  "data": {
//	    "eventId": "...",
//	    "eventType": "batch.created",
//	    "payload": { ...BatchCreateResponse... }   <-- đây là data thực
//	  }
//	}
type batchEventEnvelope struct {
	Pattern string `json:"pattern"`
	Data    struct {
		EventID  string          `json:"event_id"`
		Payload  json.RawMessage `json:"payload"` // BatchCreateResponse được nest ở đây
	} `json:"data"`
}

type ProductItemService interface {
	CreateProductItem(ctx context.Context, batch *batchDTO.BatchCreateResponse) (*response.ProductItemCreateResponse, error)
	ConsumeBatchEvent(ctx context.Context) error
	GetProductItemList(ctx context.Context, req *request.GetProductItemListRequest) (*response.ProductItemListResponse, error)
	GetProductItemDetail(ctx context.Context, itemCode string) (*response.ProductItemDetailDTO, error)
}

type productItemService struct {
	repo        repositories.ProductItemRepository
	batchRepo   batchRepo.BatchRepository
	variantRepo variantRepo.ProductVariantRepository
	consumer    *consumer.Consumer
}

func NewProductItemService(repo repositories.ProductItemRepository, batchRepo batchRepo.BatchRepository, variantRepo variantRepo.ProductVariantRepository, consumer *consumer.Consumer) ProductItemService {
	return &productItemService{
		repo:        repo,
		batchRepo:   batchRepo,
		variantRepo: variantRepo,
		consumer:    consumer,
	}
}

// CreateProductItem tạo một product item mới.
// Các trường định danh (item_code, serial_number, verification_token) được
// tự động sinh tại tầng repository theo format đã quy định:
//
//   - item_code:          PTA-{YYMM}-{8 ký tự HEX viết hoa}   regex: ^PTA-\d{4}-[A-F0-9]{8}$
//   - serial_number:      SN{14 chữ số ngẫu nhiên}             regex: ^SN\d{14}$
//   - verification_token: MD5 hex 32 ký tự thường              regex: ^[a-f0-9]{32}$
func (s *productItemService) CreateProductItem(ctx context.Context, batch *batchDTO.BatchCreateResponse) (*response.ProductItemCreateResponse, error) {
	// FK check: batch_id phải tồn tại trong bảng batches (chỉ khi được truyền vào).
	if batch.ID != uuid.Nil {
		batchExists, err := s.batchRepo.ExistsByID(ctx, batch.ID)
		if err != nil {
			return nil, err
		}
		if !batchExists {
			return nil, apperror.NewNotFound("batch")
		}
	}

	// FK check: variant_id phải tồn tại trong bảng product_variants.
	variantExists, err := s.variantRepo.ExistsByID(ctx, batch.VariantID)
	if err != nil {
		return nil, err
	}
	if !variantExists {
		return nil, apperror.NewNotFound("product_variant")
	}

	items := make([]entities.ProductItem, 0)

	for i := 0; i < batch.Quantity; i++ {
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

		pi := mapper.CreateProductItemRequestToEntity(newID, itemCode, verificationToken, serialNumber, batch)
		items = append(items, *pi)
	}

	if err := s.repo.Create(ctx, items); err != nil {
		return nil, apperror.NewInternal("failed to create product item")
	}

	return &response.ProductItemCreateResponse{
		Quantity: len(items),
	}, nil
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

func (s *productItemService) ConsumeBatchEvent(ctx context.Context) error {
	// Bug #1 fixed: dùng tên queue "batch.events" thay vì routing key BatchCreatedRK ("batch.created").
	// Bug #2 fixed: publisher.Publish() wrap message theo NestJS format {"pattern": "...", "data": {...}},
	//               nên cần unwrap trước khi unmarshal vào BatchCreateResponse.
	// Bug #3 fixed: bỏ msg.Ack/Nack trong handler — consumer.dispatch đã quản lý acknowledgment;
	//               gọi 2 lần sẽ gây PRECONDITION_FAILED (unknown delivery tag) từ RabbitMQ.
	err := s.consumer.StartConsumer(&consumer.ConsumerSpec{
		Queue:    "batch.events",
		Prefetch: 1,
		Handler: func(ctx context.Context, msg amqp.Delivery) error {
			// Step 1: Unwrap outer envelope {pattern, data: types.Event}
			var envelope batchEventEnvelope
			if err := json.Unmarshal(msg.Body, &envelope); err != nil {
				log.Printf("[BatchWorker] Failed to unmarshal envelope: %v | body: %s", err, string(msg.Body))
				return err
			}

			// Step 2: Unmarshal payload (BatchCreateResponse) từ data.payload
			var batch batchDTO.BatchCreateResponse
			if err := json.Unmarshal(envelope.Data.Payload, &batch); err != nil {
				log.Printf("[BatchWorker] Failed to unmarshal batch payload: %v | payload: %s", err, string(envelope.Data.Payload))
				return err
			}

			log.Printf("[BatchWorker] Processing batch event: id=%v, quantity=%d", batch.ID, batch.Quantity)

			_, err := s.CreateProductItem(ctx, &batch)
			if err != nil {
				log.Printf("[BatchWorker] Failed to create product item for batch %v: %v", batch.ID, err)
				return err
			}
			log.Printf("[BatchWorker] Batch event processed successfully: %v", batch.ID)
			return nil
		},
	})
	if err != nil {
		return apperror.NewInternal("failed to consume batch event")
	}

	return nil
}

func (s *productItemService) GetProductItemList(ctx context.Context, req *request.GetProductItemListRequest) (*response.ProductItemListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	return s.repo.FindAllWithFilter(ctx, req)
}

func (s *productItemService) GetProductItemDetail(ctx context.Context, itemCode string) (*response.ProductItemDetailDTO, error) {
	return s.repo.FindByItemCodeWithEvents(ctx, itemCode)
}
