# BÁO CÁO ĐÁNH GIÁ KIẾN TRÚC HỆ THỐNG (ARCHITECTURAL AUDIT REPORT)
## Dự án: ProductTrace AI
**Vai trò:** Senior Solution Architect & Staff Engineer
**Ngày thực hiện:** 16-06-2026

---

## 📊 1. Đánh giá tổng quan (Overall Audit Summary)

* **Điểm đánh giá kiến trúc hiện tại:** **4 / 10** (Chưa sẵn sàng cho Production - *Not Production Ready*)
* **Nhận định chính:** Hệ thống hiện tại có một bản thiết kế nghiệp vụ (blueprint) rất triển vọng với việc tận dụng tối đa sức mạnh của **PostgreSQL + PostGIS** cho dữ liệu không gian hành trình và tách biệt domain giữa **Go Core** (transactional/business logic) và **NestJS** (AI/Search/Vector database). Tuy nhiên, phần triển khai thực tế mới dừng lại ở mức **khung thô (skeleton/boilerplate)**. Hệ thống chứa nhiều **anti-pattern chí tử** có thể gây sập hệ thống (blocking event loop), mất mát dữ liệu sự kiện (event loss), cấu hình Docker chưa bảo mật và định tuyến Gateway chưa hoàn thiện.

---

## 🏛️ 2. Đánh giá chi tiết theo 10 Tiêu chí yêu cầu

### 2.1. Kiến trúc tổng thể (Overall Architecture)
* **Khớp nối chặt chẽ (Tight Coupling)**: Nest AI thực hiện gọi REST đồng bộ ngược sang Go Core thông qua biến môi trường `GO_CORE_URL`. Đây là dấu hiệu của **Distributed Monolith**. Nếu Go Core bị chậm hoặc sập, luồng xử lý của Nest AI sẽ lập tức bị gián đoạn.
* **Hướng khắc phục**: 
  * Chuyển đổi các giao tiếp đồng bộ bắt buộc sang **gRPC** để tối ưu hóa serialization, giảm latency.
  * Các luồng nghiệp vụ không cần phản hồi tức thì (như cập nhật trạng thái, đồng bộ log hành trình) phải được chuyển đổi hoàn toàn sang **Event-Driven Architecture (EDA)** bất đồng bộ thông qua RabbitMQ.

### 2.2. Monorepo
* **Hiện trạng**: Cấu trúc thư mục tương đối mạch lạc với phân chia rõ ràng (`apps/go-core-service`, `apps/nest-ai-service`, `packages/contracts`). Tuy nhiên, thư mục `packages/contracts` hiện tại chỉ chứa file `readme.md`, không có code hay schema thực tế.
* **Hướng khắc phục**:
  * Tránh việc định nghĩa lại cấu trúc Event Payload thủ công ở cả hai phía Go và NestJS (dễ gây ra trôi lệch schema - *Schema Drift*).
  * Sử dụng **Protocol Buffers (Protobuf)** làm Single Source of Truth cho toàn bộ API DTOs và Event Schemas tại thư mục `packages/contracts`. Dùng code generator để tự động xuất ra Go Structs và TypeScript Interfaces/DTOs.

### 2.3. RabbitMQ Topology & Reliability
* **Vấn đề Topology**: Chỉ định nghĩa 1 Queue duy nhất `ai.events` nhận các event. Khi hệ thống mở rộng lên 10+ services, việc dùng chung một hàng đợi sẽ gây nghẽn cổ chai. Nguyên tắc bất di bất dịch: **Mỗi service phải có queue riêng biệt**.
* **Lỗi xử lý sự kiện (Event Loss)**: Go Core thực hiện ghi DB rồi publish thẳng sang RabbitMQ. Nếu broker bị mất kết nối ngay sau khi DB commit, event sẽ bị mất vĩnh viễn.
* **Thiếu Retry Queue**: Consumer Nest AI khi lỗi xử lý sẽ `nack(message, false, false)` để đẩy thẳng message lỗi sang DLQ (`ai.events.failed`). Thiếu cơ chế tự động thử lại (Exponential Backoff Retry).
* **Hướng khắc phục**:
  * Triển khai **Transactional Outbox Pattern** ở Go Core.
  * Cấu hình **Retry Queues** với thuộc tính TTL (Time To Live) và Dead Letter Exchange (DLX) để tự động retry message 3-5 lần trước khi chuyển hẳn vào DLQ.

### 2.4. Định tuyến & Bảo mật tại Kong Gateway
* **Hiện trạng**: Cấu hình Kong tĩnh (`kong.yml`) ở chế độ DB-less rất tốt về hiệu năng, nhưng cấu hình định tuyến hiện tại cực kỳ cơ bản và thiếu các lớp bảo vệ/giám sát thiết yếu cho production.
* **Hướng khắc phục**: Kích hoạt và cấu hình các plugin production:
  * **JWT Plugin**: Xác thực JWT tập trung tại Gateway. Các service phía sau chỉ nhận thông tin định danh đáng tin cậy qua Header `X-User-Id`.
  * **Rate Limiting**: Giới hạn tần suất gọi API (theo IP hoặc User ID) để chống DDoS.
  * **Correlation ID**: Tự động sinh trace ID (`X-Correlation-Id`) cho mọi request đầu vào để phục vụ distributed tracing.
  * **CORS Plugin**: Cấu hình CORS tập trung thay vì cấu hình rải rác ở code app.

### 2.5. Go Core Service
* **Hiện trạng**: Tệp `main.go` hiện tại được viết dưới dạng kịch bản thử nghiệm nhanh, thiếu hoàn toàn cấu trúc Clean Architecture thực tế, thiếu Dependency Injection (DI), sử dụng package `log` mặc định của Go, và đọc trực tiếp biến môi trường bằng `os.Getenv` rải rác.
* **Hướng khắc phục**:
  * Phân rã dự án thành 3 layer chuẩn: **Domain** (Entities/Interfaces), **Usecase** (Business logic), và **Adapter** (HTTP controller, DB repository, RabbitMQ publisher).
  * Sử dụng thư viện quản lý cấu hình tập trung (như **Viper** hoặc **Cleanenv**) kèm cơ chế validate schema lúc khởi chạy.
  * Thay thế package `log` bằng **Structured Logging** (`slog` hoặc `zap`) để ghi log dưới dạng JSON chuẩn hóa.
  * Sử dụng **Fx (Uber)** hoặc **Wire (Google)** để làm Dependency Injection sạch sẽ khi số lượng service tăng lên.

### 2.6. Nest AI Service
* **Vấn đề nghẽn Event Loop**: Xử lý AI/Embedding là các tác vụ CPU-bound cực nặng. Nếu xử lý trực tiếp trên event loop đơn luồng của Node.js bên trong consumer handler, nó sẽ block toàn bộ việc nhận và xử lý các sự kiện tiếp theo.
* **Hướng khắc phục**:
  * Khi nhận event từ RabbitMQ, Consumer chỉ làm nhiệm vụ ghi nhận và đẩy nhanh tác vụ vào hàng đợi job nền (**BullMQ** dựa trên Redis).
  * Triển khai cụm **AI Workers** riêng biệt (hoặc sử dụng **Node.js Worker Threads**) để tiêu thụ job từ BullMQ, giải phóng Main Thread của NestJS.

### 2.7. Docker & Infrastructure
* **Thiếu tài nguyên cốt lõi**: File `docker-compose.yml` thiếu hoàn toàn dịch vụ **Qdrant Vector DB** - hạ tầng bắt buộc cho tìm kiếm ngữ nghĩa của Nest AI.
* **Dockerfile chưa tối ưu**: Dockerfile của Nest AI không dùng multi-stage build làm dung lượng phình to (>1GB) và lộ source code TS. Dockerfile của Go Core phá vỡ cache layer khi copy source code trước khi chạy `go mod download`. Cả hai container đều đang chạy với quyền hạn cao nhất (`root`).
* **Hạ tầng Redis**: Đang sử dụng tag `redis:8-alpine` (Redis 8 hiện vẫn ở giai đoạn phát triển thử nghiệm/alpha). Redis không cấu hình mật khẩu bảo mật.

### 2.8. Observability (Giám sát hệ thống)
* **Hiện trạng**: Hệ thống hoàn toàn trống rỗng về mặt đo lường (Metrics), giám sát vết (Tracing), và log cấu trúc.
* **Hướng khắc phục**:
  * Tích hợp **OpenTelemetry SDK** vào cả Go Core và Nest AI.
  * Đọc và truyền tiếp `X-Correlation-Id` qua Header HTTP và RabbitMQ Message Headers.
  * Thiết lập cụm **Prometheus + Grafana** để giám sát tài nguyên và hàng đợi RabbitMQ.

### 2.9. Bảo mật (Security)
* **Secrets Management**: Tuyệt đối không hardcode mật khẩu database, Redis, RabbitMQ trong docker-compose file. Cần chuyển dịch sang dùng tham chiếu biến môi trường qua file `.env`.
* **Mạng nội bộ (Private Network)**: Kín cổng cao tường cho các dịch vụ phía sau. Chỉ expose duy nhất port `8000` (Kong Proxy) ra ngoài. Các service khác chỉ giao tiếp nội bộ trong docker network.

### 2.10. Khả năng mở rộng (Scale-out)
* Khi tăng lên 10+ services, 100.000+ events/ngày:
  * RabbitMQ cần cấu hình cụm (Clustering) kết hợp **Quorum Queues** để bảo vệ dữ liệu khỏi phân mảnh mạng (Network Partition).
  * Outbox Pattern và Retry Queue trở thành bắt buộc để bảo vệ tính toàn vẹn dữ liệu.
  * Các AI Workers cần được scale độc lập (ví dụ sử dụng HPA trên Kubernetes dựa trên độ dài hàng đợi của BullMQ thay vì CPU/Memory).

---

## 🗺️ 3. Sơ đồ Kiến trúc cải tiến đề xuất (Target Architecture)

```mermaid
graph TD
    Client[Client Apps / Web] -->|HTTPS| Gateway[Kong API Gateway - Declarative DB-less]
    
    subgraph Infrastructure
        RabbitMQ[RabbitMQ Message Broker]
        Redis[Redis Shared Cache v7]
        PostgreSQL[(PostgreSQL + PostGIS)]
        Qdrant[(Qdrant Vector Database)]
    end

    subgraph Go Core Services
        Gateway -->|Route /api| GoCore[Go Core Service - Non-root]
        GoCore -->|Write Transaction| PostgreSQL
        GoCore -->|Write Outbox Table| PostgreSQL
        GoCore -->|Cache State| Redis
        
        GoOutbox[Outbox Worker] -->|Read Outbox & Publish| RabbitMQ
    end

    subgraph Nest AI Services
        Gateway -->|Route /ai| NestAI[NestJS API Gateway / Consumer]
        NestAI -->|Query Semantic| Qdrant
        NestAI -->|gRPC / Cache Read| GoCore
        
        RabbitMQ -->|Consume Event| NestAI
        NestAI -->|Enqueues Job| BullMQ[(BullMQ / Redis)]
        BullMQ -->|Process CPU-bound AI| AIWorker[Node.js AI Worker Threads]
        AIWorker -->|Upsert Embeddings| Qdrant
    end
```

---

## 🛠️ 4. Chi tiết các Anti-patterns & Giải pháp Code/Config mẫu

### 4.1. Khắc phục lỗi thiếu trường Database & Thiết lập giá trị mặc định
Sửa đổi file migration khởi tạo database (`migrations/000001_init_database_schema.up.sql`) để bổ sung các cột dữ liệu bảo mật, check constraints định dạng mã vật lý và giá trị mặc định cho UUID/Timestamps.

```diff
-- Bảng product_items bổ sung verification_token và ràng buộc định dạng
CREATE TABLE "product_items" (
- "id" uuid PRIMARY KEY,
+ "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "variant_id" uuid NOT NULL,
  "batch_id" uuid,
  "current_shop_point_id" uuid,
- "item_code" varchar UNIQUE NOT NULL,
+ "item_code" varchar(20) UNIQUE NOT NULL,
+ "verification_token" varchar(32) UNIQUE NOT NULL,
  "serial_number" varchar UNIQUE,
  "status" varchar DEFAULT 'IN_STOCK',
  "produced_at" timestamp,
  "packed_at" timestamp,
  "sold_at" timestamp,
  "registered_at" timestamp,
  "last_scanned_at" timestamp,
  "metadata_json" jsonb,
- "created_at" timestamp,
- "updated_at" timestamp,
+ "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
+ "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "is_deleted" boolean DEFAULT false
);

+ -- Thêm check constraints định dạng mã vật lý và token
+ ALTER TABLE product_items ADD CONSTRAINT chk_product_items_item_code_format CHECK (item_code ~ '^PTA-[0-9]{4}-[A-Z0-9]{8}$');
+ ALTER TABLE product_items ADD CONSTRAINT chk_product_items_verification_token_format CHECK (verification_token ~ '^[a-f0-9]{32}$');
+ CREATE INDEX idx_product_items_code_token ON product_items (item_code, verification_token);

-- Đổi tên bảng nhật ký đồng bộ snake_case
- CREATE TABLE "auditLog" (
+ CREATE TABLE "audit_logs" (
-  "id" uuid PRIMARY KEY,
+  "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
   "content" text NOT NULL,
   "type" varchar,
-  "created_at" timestamp
+  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP
 );
```

---

### 4.2. Triển khai Transactional Outbox Pattern ở Go Core
Đảm bảo khi có thay đổi trạng thái nghiệp vụ (ví dụ: tạo sản phẩm), dữ liệu chính và sự kiện được ghi vào DB trong cùng một Transaction để chống mất mát event.

#### Bảng lưu trữ Outbox Events:
```sql
CREATE TABLE "outbox_events" (
  "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "event_type" varchar(100) NOT NULL,
  "payload" jsonb NOT NULL,
  "status" varchar(20) DEFAULT 'PENDING', -- PENDING, PROCESSED, FAILED
  "correlation_id" varchar(50),
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "processed_at" timestamp
);
CREATE INDEX idx_outbox_events_status ON outbox_events(status);
```

#### Go implementation (Usecase Layer):
```go
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
)

type ProductUsecase struct {
	db *sql.DB
}

func NewProductUsecase(db *sql.DB) *ProductUsecase {
	return &ProductUsecase{db: db}
}

func (u *ProductUsecase) CreateProduct(ctx context.Context, name string, categoryID string) error {
	// Khởi chạy Database Transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	productID := uuid.New().String()
	// 1. Lưu sản phẩm mới
	_, err = tx.ExecContext(ctx, 
		`INSERT INTO products (id, category_id, name, status) VALUES ($1, $2, $3, 'ACTIVE')`,
		productID, categoryID, name,
	)
	if err != nil {
		return err
	}

	// 2. Đóng gói payload sự kiện
	payload := map[string]interface{}{
		"productId": productID,
		"name":      name,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 3. Ghi sự kiện vào bảng Outbox để gửi đi sau (bất đồng bộ)
	_, err = tx.ExecContext(ctx, 
		`INSERT INTO outbox_events (event_type, payload, correlation_id) VALUES ($1, $2, $3)`,
		"product.created", payloadBytes, uuid.New().String(),
	)
	if err != nil {
		return err
	}

	// Commit Transaction - dữ liệu nghiệp vụ và sự kiện lưu thành công đồng thời
	return tx.Commit()
}
```

---

### 4.3. Thiết lập RabbitMQ Backoff Retry Queue
Sử dụng các hàng đợi trung gian có thiết lập TTL và Dead Letter Exchange để tự động thử lại khi consumer gặp sự cố tạm thời, tránh đẩy trực tiếp sang DLQ.

```go
// apps/go-core-service/internal/events/rabbitmq/topology.go
package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func SetupTopology(ch *amqp.Channel) error {
	// 1. Exchange chính (Topic)
	if err := ch.ExchangeDeclare("product-trace.events", "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// 2. Retry Exchange (Direct)
	if err := ch.ExchangeDeclare("product-trace.retry-exchange", "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// 3. Dead Letter Exchange cuối cùng (Direct)
	if err := ch.ExchangeDeclare("product-trace.dlx", "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// 4. Queue chính trỏ DLX sang Retry Exchange
	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    "product-trace.retry-exchange",
		"x-dead-letter-routing-key": "ai.events.retry",
	}
	if _, err := ch.QueueDeclare("ai.events", true, false, false, false, mainArgs); err != nil {
		return err
	}
	if err := ch.QueueBind("ai.events", "product.created", "product-trace.events", false, nil); err != nil {
		return err
	}

	// 5. Retry Queue - Trực tiếp trì hoãn 5 giây (TTL) rồi gửi trả lại Exchange chính
	retryArgs := amqp.Table{
		"x-dead-letter-exchange":    "product-trace.events",
		"x-dead-letter-routing-key": "product.created",
		"x-message-ttl":             int32(5000), // delay 5000ms (5 giây)
	}
	if _, err := ch.QueueDeclare("ai.events.retry.5s", true, false, false, false, retryArgs); err != nil {
		return err
	}
	if err := ch.QueueBind("ai.events.retry.5s", "ai.events.retry", "product-trace.retry-exchange", false, nil); err != nil {
		return err
	}

	// 6. DLQ cuối cùng dành cho các message lỗi vĩnh viễn
	if _, err := ch.QueueDeclare("ai.events.dlq", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind("ai.events.dlq", "ai.events.failed", "product-trace.dlx", false, nil); err != nil {
		return err
	}

	return nil
}
```

---

### 4.4. Cấu hình sản xuất Kong Gateway với các Plugins thiết yếu
```yaml
# infra/kong/kong.yml
_format_version: "3.0"

services:
  - name: go-core-service
    url: http://go-core-service:8080
    routes:
      - name: go-core-route
        paths:
          - /api
        strip_path: true
    plugins:
      - name: correlation-id
        config:
          header_name: X-Correlation-ID
          generator: uuid
          echo_back: true
      - name: rate-limiting
        config:
          second: 20
          hour: 10000
          policy: local
      - name: cors
        config:
          origins:
            - "*"
          methods:
            - GET
            - POST
            - PUT
            - DELETE
            - OPTIONS

  - name: nest-ai-service
    url: http://nest-ai-service:3000
    routes:
      - name: nest-ai-route
        paths:
          - /ai
        strip_path: true
    plugins:
      - name: correlation-id
        config:
          header_name: X-Correlation-ID
          generator: uuid
          echo_back: true
      - name: rate-limiting
        config:
          second: 5
          hour: 2000
          policy: local
```

---

### 4.5. Thiết kế Dockerfile Đa tầng (Multi-stage) Bảo mật & Tối ưu Cache

#### Go Core Service Dockerfile (`apps/go-core-service/Dockerfile`):
```dockerfile
# --- Stage 1: Build ---
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Sao chép files dependency trước để tận dụng cache của Docker
COPY go.mod go.sum ./
RUN go mod download

# Sao chép mã nguồn và compile tĩnh
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api ./cmd/api

# --- Stage 2: Runner ---
FROM alpine:3.22
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/api .

# Tạo user phi-root để vận hành an toàn
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080
CMD ["./api"]
```

#### Nest AI Service Dockerfile (`apps/nest-ai-service/Dockerfile`):
```dockerfile
# --- Stage 1: Builder ---
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# --- Stage 2: Production Installer ---
FROM node:22-alpine AS installer
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# --- Stage 3: Runner ---
FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=installer /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package.json ./package.json

# Chạy container bằng user 'node' hạn chế đặc quyền
USER node

EXPOSE 3000
CMD ["node", "dist/main.js"]
```

---

### 4.6. File Docker Compose hoàn thiện cho môi trường sản xuất (`docker-compose.yml`)
Bao gồm: Bổ sung **Qdrant**, hạ cấp **Redis**, bổ sung các tham số **Healthcheck**, đặt mật khẩu và loại bỏ hardcode qua `.env`.

```yaml
version: '3.8'

services:
  postgres:
    image: postgis/postgis:16-3.4
    container_name: product-trace-db
    environment:
      POSTGRES_DB: ${DB_NAME:-product_trace_db}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    ports:
      - "5433:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - app-networks

  redis:
    image: redis:7-alpine
    container_name: product-trace-redis
    command: redis-server --requirepass ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - app-networks

  rabbitmq:
    image: rabbitmq:3.13-management
    container_name: product-trace-rabbitmq
    environment:
      RABBITMQ_DEFAULT_USER: ${RABBITMQ_USER:-admin}
      RABBITMQ_DEFAULT_PASS: ${RABBITMQ_PASSWORD}
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq-data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "check_running"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-networks

  qdrant:
    image: qdrant/qdrant:v1.9.0
    container_name: product-trace-qdrant
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant-data:/qdrant/storage
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:6333/healthz"]
      interval: 10s
      timeout: 5s
      retries: 3
    networks:
      - app-networks

  migrate:
    image: migrate/migrate
    container_name: product-trace-migrate
    volumes:
      - ./migrations:/migrations:ro
    command:
      [
        "-path",
        "/migrations",
        "-database",
        "postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable",
        "up"
      ]
    restart: on-failure
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - app-networks

  go-core-service:
    build:
      context: ./apps/go-core-service
      dockerfile: Dockerfile
    container_name: go-core-service
    environment:
      DB_HOST: postgres
      DB_PORT: "5432"
      DB_NAME: ${DB_NAME}
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      REDIS_HOST: redis
      REDIS_PORT: "6379"
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      RABBITMQ_URL: amqp://${RABBITMQ_USER}:${RABBITMQ_PASSWORD}@rabbitmq:5672/
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 5s
      timeout: 3s
      retries: 3
    depends_on:
      migrate:
        condition: service_completed_successfully
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    networks:
      - app-networks

  nest-ai-service:
    build:
      context: ./apps/nest-ai-service
      dockerfile: Dockerfile
    container_name: nest-ai-service
    environment:
      PORT: 3000
      GO_CORE_URL: http://go-core-service:8080
      REDIS_HOST: redis
      REDIS_PORT: "6379"
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      RABBITMQ_URL: amqp://${RABBITMQ_USER}:${RABBITMQ_PASSWORD}@rabbitmq:5672/
      RABBITMQ_QUEUE: ai.events
      QDRANT_HOST: qdrant
      QDRANT_PORT: "6333"
    ports:
      - "3000:3000"
    depends_on:
      go-core-service:
        condition: service_healthy
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      qdrant:
        condition: service_healthy
    networks:
      - app-networks

networks:
  app-networks:
    name: app-networks
    driver: bridge

volumes:
  postgres-data:
  rabbitmq-data:
  qdrant-data:
```

---

## 📅 5. Kế hoạch hành động theo độ ưu tiên (Execution Roadmap)

### 🔴 Giai đoạn 1: Sửa lỗi Core & Hạ tầng (P0 - Khẩn cấp trong tuần 1)
1. **Database Schema Alignment**: Sửa đổi file migration, bổ sung các check constraints và cột `verification_token`.
2. **Docker Optimization**: Cập nhật lại 2 Dockerfile theo dạng multi-stage build, hạ cấp Redis xuống v7 và thêm cấu hình Qdrant DB vào compose.
3. **Healthcheck & Boot Order**: Thiết lập lại thuộc tính `depends_on` kèm điều kiện `service_healthy` để hệ thống khởi động trơn tru không bị crash-loop.

### 🟡 Giai đoạn 2: Gia cố Độ tin cậy & Giao tiếp (P1 - Tuần 2-3)
1. **Outbox Pattern**: Tạo bảng `outbox_events` và viết background publisher trong Go Core để chống mất mát sự kiện.
2. **Retry Queues**: Tái thiết kế topology của RabbitMQ trong Go Core, chuyển các queue về dạng có cấu hình Retry Backoff thay vì dội thẳng vào DLQ.
3. **Kong Plugins**: Thêm cấu hình JWT authentication, Rate limiting và Correlation ID tại lớp Gateway.
4. **BullMQ Integration**: Cài đặt BullMQ trong Nest AI để chạy các job tính toán embedding nặng, tránh block event loop Node.js.

### 🔵 Giai đoạn 3: Mở rộng, Bảo mật & Giám sát (P2 - Tuần 4+)
1. **Protobuf Schema Sharing**: Khai báo tệp định nghĩa Protobuf dùng chung tại `packages/contracts` để biên dịch sang Go Struct và Nest TS Interfaces.
2. **OpenTelemetry Setup**: Cấu hình logging JSON, thêm OpenTelemetry SDK để theo dõi trace ID xuyên suốt toàn hệ thống.
3. **mTLS/Internal Keys**: Bảo mật giao tiếp nội bộ giữa các microservices.

---

## 📋 6. Checklist Production Readiness (Vận hành thực tế)

- [ ] **Docker Security**: Các container hoàn toàn chạy bằng user hạn chế đặc quyền (non-root).
- [ ] **Infrastructure Secrets**: Không còn bất kỳ mật khẩu nào bị hardcode dưới dạng plain-text trong Docker Compose hoặc code.
- [ ] **Database Integrity**: Bảng `product_items` có đầy đủ các check constraints mã hóa vật lý và `verification_token`.
- [ ] **Database Indexes**: Tạo thành công composite index `idx_product_items_code_token` và các Spatial Indexes (`GIST`) cho địa lý.
- [ ] **Event Delivery Guarantee**: Đã kích hoạt Outbox Pattern ở Go Core và message broker lưu trữ ở dạng durable (persistent).
- [ ] **Topology Isolation**: Mỗi microservice đăng ký queue độc lập, không dùng chung queue tiêu thụ sự kiện.
- [ ] **No Blocking Event Loop**: Tác vụ CPU-bound trong Nest AI được đẩy hoàn toàn sang BullMQ worker.
- [ ] **API Security**: Kong Gateway thực thi giới hạn lưu lượng (Rate limiting) và chặn truy cập trực tiếp vào microservices từ ngoài mạng ảo.
- [ ] **Tracing Visibility**: Trace ID truyền thành công qua header HTTP và header RabbitMQ để debug luồng nghiệp vụ.
- [ ] **Graceful Shutdown**: Cả Go và NestJS đều bắt các tín hiệu `SIGTERM`/`SIGINT` để đóng kết nối DB, Redis và hoàn thành xử lý message đang dang dở trước khi thoát.
