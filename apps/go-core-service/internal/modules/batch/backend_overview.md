# Backend Overview — go-core-service

> **Purpose**: Một file, đủ context để hiểu toàn bộ backend mà không cần đọc source.  
> **Module path**: `github.com/khangpd15/producttrace-ai/apps/go-core-service`  
> **Go version**: 1.25

---

## 1. Architecture

### Framework & Libraries
| Concern | Library |
|---|---|
| HTTP Framework | `gin-gonic/gin v1.12` |
| ORM | `gorm.io/gorm v1.31` + PostgreSQL driver |
| Database | PostgreSQL (via `gorm.io/driver/postgres`) |
| Cache | Redis (`redis/go-redis/v9`) |
| Message Queue | RabbitMQ (`rabbitmq/amqp091-go`) |
| JWT | `golang-jwt/jwt/v5` |
| QR code | `skip2/go-qrcode` |
| PDF generation | `go-pdf/fpdf` |
| Excel generation | `xuri/excelize/v2` |
| Env | `joho/godotenv` |

### Project Structure
```
cmd/
  api/         ← HTTP server entrypoint
  worker/      ← Worker entrypoint (placeholder, no active workers)
internal/
  app/         ← NewApp() — manual DI wiring
  router/      ← SetupRouter() + per-module route functions
  middleware/  ← auth, role, rate_limiter, logger, recovery, request_id
  events/
    publisher/ ← RabbitMQ publish wrapper
    consumer/  ← RabbitMQ consume wrapper
    rabbitmq/  ← Manager, topology, constants
    types/     ← Event struct
  modules/
    authen/    ← Login, Register, OTP, Refresh, Forgot/Reset password
    user/      ← CRUD user, profile
    batch/     ← Batch management (primary module)
    trace/     ← Product traceability timeline
    event/     ← Event lifecycle (create only)
    product/   ← Product CRUD
    product_variant/
    product_attribute/
    product_attribute_value/
    product_item/  ← QR scan, product item CRUD
    location/  ← Location CRUD
    public/    ← Anonymous QR verification
  workers/     ← Empty (readme only)
  realtime/    ← (exists, not wired into main flow)
pkg/
  apperror/    ← AppError type + HTTP error mapping
  audit_log/   ← Shared audit log service (model, repo, service)
  cache/       ← Redis cache wrapper
  response/    ← Standardised ResponseSuccess / ResponseError envelope
```

### Dependency Injection
- Manual constructor wiring in [`app.go`](file:///d:/Workspace/projects/product-trace-ai/src-code/BE/producttrace-ai/apps/go-core-service/internal/app/app.go): `NewApp(db, redis, publisher, sqlDB)`.
- No DI framework. Each module wires: `Repository → Service → Handler`.
- Shared `auditService` is injected into every module that needs audit logging.

### Middleware Stack (global)
| Middleware | Scope |
|---|---|
| `RecoveryMiddleware` | Global — panic recovery |
| `RequestIDMiddleware` | Global — injects X-Request-ID |
| `LoggerMiddleware` | Global — request logging |
| `AuthMiddleware(userRepo)` | Route groups requiring auth — validates JWT, sets `user_id`, `email`, `role` in context |
| `RoleMiddleware(roles...)` | Route groups requiring specific roles |
| `RateLimiter.Limit(n, period)` | Trace search endpoint — Redis-backed, in-memory fallback |

> CORS is **disabled** in Go code — handled externally by Kong Gateway.

### Routing
- Base prefix: `/api`
- All route groups defined in [`router.go`](file:///d:/Workspace/projects/product-trace-ai/src-code/BE/producttrace-ai/apps/go-core-service/internal/router/router.go)
- Static file serving: `/storage → ./storage`

### Message Queue (RabbitMQ)
- Exchange: `product-trace.events` (topic exchange)
- DLX: `product-trace.dlx`
- Publisher: fire-and-forget via `publisher.Publish(Event)`. Non-blocking.
- Consumer: `consumer.StartConsumer(spec)` — dispatches via `go dispatch(...)` with 30s timeout per message, nack+requeue on error.
- Routing keys declared: `user.*`, `product.*`, `batch.*`, `trace.*`, `otp.*`, `notification.*`

### Worker
- `cmd/worker/` exists but the `internal/workers/` directory has only an empty readme. No active background workers are implemented.

---

## 2. Module Batch

### Purpose
Quản lý lô hàng sản xuất/nhập khẩu. Là trung tâm kết nối giữa Product Variant → Product Items → Events.

### API

| Method | Endpoint | Auth | Role |
|---|---|---|---|
| GET | `/api/batches/:id` | Public | — |
| GET | `/api/batches` | JWT | Any |
| GET | `/api/batches/search` | JWT | Any |
| GET | `/api/batches/:id/events` | JWT | Any |
| GET | `/api/batches/:id/products` | JWT | Any |
| GET | `/api/batches/:id/history` | JWT | ADMIN, STAFF |
| GET | `/api/batches/export-qr/:id` | JWT | ADMIN, MANUFACTURER |
| POST | `/api/batches` | JWT | ADMIN, MANUFACTURER |
| PATCH | `/api/batches/:id/status` | JWT | ADMIN, MANUFACTURER |
| DELETE | `/api/batches/:id` | JWT | ADMIN, MANUFACTURER |
| POST | `/api/batches/:id/export` | JWT | ADMIN, MANAGER, WAREHOUSE |

### Request DTOs
| DTO | Key Fields |
|---|---|
| `CreateBatchRequest` | `variant_id`, `prefix` (2–20 alpha), `manufacture_date`, `expiry_date`, `quantity` |
| `ExportBatchRequest` | `destination_location`, `quantity`, `operator_name`, `notes` |
| `UpdateBatchStatusRequest` | `status` |
| `GetBatchListRequest` | `search`, `status`, `origin_country`, page/limit |
| `SearchBatchRequest` | `keyword` (max 100), `sort_by`, `sort_order`, page/limit |
| `GetBatchProductsRequest` | `page` (default 1), `limit` (default 20, max 100), `status`, `keyword` (max 50) |
| `GetBatchHistoryRequest` | `page`, `limit` |

### Response DTOs
| DTO | Description |
|---|---|
| `BatchCreateResponse` | `id`, `batch_code`, `variant_id`, `quantity`, `status`, `created_at` |
| `BatchDetailResponse` | Full detail including variant/product info |
| `BatchListResponse` | Paginated list + `stats` (counts by status) |
| `SearchBatchResponse` | Paginated search results |
| `BatchStatusResponse` | `id`, `batch_code`, `status`, `updated_at` |
| `GetBatchProductsResponse` | `items[]` (itemId, itemCode, serialNumber, status, currentLocation, createdAt) + `pagination` |
| `GetBatchHistoryResponse` | `batchId`, `batchCode`, `history[]` (logId, action, changedFields diff, performedBy, createdAt) |
| `BatchEventDTO` | `event_name`, `detail`, `created_at` |

### Service Flow

**CreateBatch**:
1. Normalize prefix (uppercase).
2. Validate `variant_id` exists (`variantRepo.ExistsByID`).
3. Validate `expiry_date >= manufacture_date`, `manufacture_date` not in future, `quantity > 0`.
4. Call `repo.Create()` with retry (3 attempts, 100ms backoff — for transient DB errors only).
5. Repo acquires PostgreSQL advisory lock (`pg_advisory_xact_lock(hashtext(prefix-year))`), finds max `batch_code`, increments sequence, INSERT.
6. Publish `batch.created` event to RabbitMQ.

**UpdateBatchStatus**:
1. Validate status enum: `{ACTIVE, EXPIRED, RECALLED, BLOCKED}`.
2. `FindByID` — error if not found or already soft-deleted.
3. `repo.UpdateStatus`.
4. Write audit log (`pkg/audit_log.LogUpdate`).

**DeleteBatch** (soft-delete):
1. `FindByID`.
2. Check not already deleted.
3. `ExistsProductItems` — block if any linked product items.
4. `ExistsEvents` — block if any linked events.
5. `SoftDelete`.
6. Write audit log.

**ExportBatch**:
- Runs in DB transaction: fetch batch, check `batch.Quantity >= req.Quantity`, decrement quantity, write audit log row directly to `audit_logs` table.

**ExportBatchQR**:
- Fetch batch detail + all `product_items` by `batch_id`.
- Generate PDF of QR labels using `go-qrcode` + `fpdf`. Returns raw bytes.

**GetBatchHistory**:
- Verify batch exists.
- Query `audit_logs` JOIN `users`, paginated.
- Publish `batch.history_viewed` event (fire-and-forget).

**GetBatchProducts**:
- Verify batch exists.
- Normalize status filter.
- Query `product_items` LEFT JOIN `locations`, with pagination, status whitelist filter, ILIKE search on `item_code`/`serial_number`.

### Repository
Interface: [`BatchRepository`](file:///d:/Workspace/projects/product-trace-ai/src-code/BE/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories/batch_repository.go#L22)

Methods: `FindAllWithFilter`, `SearchBatches`, `FindByCode`, `FindByBatchID`, `GetBatchEvents`, `FindByID`, `ExistsByID`, `Create`, `UpdateStatus`, `SoftDelete`, `ExportBatch`, `ExistsProductItems`, `ExistsEvents`, `GetBatchHistory`, `GetBatchProducts`.

All queries use GORM. `SearchBatches` applies sortBy whitelist (`batch_code`, `manufacture_date`, `expiry_date`, `quantity`, `created_at`).

### Database Tables Used
`batches`, `product_variants`, `products`, `product_items`, `events`, `locations`, `audit_logs`, `users`

### Business Rules
- `batch_code` format: `{PREFIX}-{YEAR}-{NNNN}` (4-digit zero-padded sequence). Auto-generated.
- Advisory lock prevents sequence race conditions per `prefix-year`.
- `DRAFT` status: only visible to ADMIN (BR-FIL-001/002).
- Valid filter statuses: `DRAFT, ACTIVE, EXPIRED, RECALLED, LOCKED`.
- Valid update statuses: `ACTIVE, EXPIRED, RECALLED, BLOCKED`.
- Cannot delete batch if it has linked product items or events.
- Batch quantity decremented on export; export logged in `audit_logs`.

### Dependencies
- `product_variant` repo (FK validation on create)
- `product_item` repo (QR export)
- `audit_log` pkg (update/delete/history logging)
- RabbitMQ publisher (batch.created, batch.history_viewed)

---

## 3. Module Trace

### Purpose
Truy xuất nguồn gốc sản phẩm theo timeline. Cho phép tra cứu toàn bộ vòng đời sự kiện của một product item bằng `item_code` hoặc `serial_number`. Hỗ trợ export PDF/Excel.

### API

| Method | Endpoint | Auth | Role / Limit |
|---|---|---|---|
| GET | `/api/trace/search` | Public | Rate limit: 30 req/min/IP |
| POST | `/api/trace/export/pdf` | JWT | Any authenticated |
| POST | `/api/trace/export/excel` | JWT | Any authenticated |

### Request DTOs
| DTO | Key Fields |
|---|---|
| `TraceSearchRequest` | `code` (required, 3–100), `fromDate` (ISO-8601), `toDate`, `eventTypes` (comma-separated) |
| `PDFExportRequest` | `productItemId` (UUID, required), `theme` (WARM_MINIMAL\|CLASSIC_NAVY), `includeAuditLogs` |
| `ExcelExportRequest` | `productItemId` OR `batchId` (at least one required), `fromDate`, `toDate` |

### Response DTOs
| DTO | Description |
|---|---|
| `TraceSearchResponse` | `productItem?`, `timeline[]`, `warning?`, `filterApplied?`, `matchedEventsCount?` |
| `ExportJobResponse` | `jobId`, `status`, `estimatedTimeSeconds` |

### Trace Search Flow (SearchTimeline)
1. Validate `code` (3–100 chars).
2. Parse and validate `fromDate`/`toDate` (RFC3339). Validate range.
3. Parse `eventTypes` comma-separated, validate against whitelist.
4. `repo.FindProductItemByCode(code)` — matches `item_code` OR `serial_number`.
5. **Role-based event filtering**:
   - Privileged (`ADMIN`, `STAFF`, `DEALER`): all event types.
   - Public / unauthenticated / `CUSTOMER`: internal event types removed (`PACKED`, `WAREHOUSE_IN`, `WAREHOUSE_OUT`, `TRANSPORTED`).
6. `repo.FindEvents(itemID, batchID, fromDate, toDate, eventTypes)`.
7. Write audit log (`PUBLIC_SEARCH_TIMELINE` / `FILTER_TIMELINE_BY_DATE` / `FILTER_TIMELINE_BY_TYPE`).
8. Add `WARNING` if `item.Status == "RECALLED"`.
9. Return timeline. If filters applied → include `filterApplied` and `matchedEventsCount`. If no filters → include `productItem` detail.

### Export PDF Flow
1. Lookup product item by UUID.
2. Fetch all events (no date/type filter).
3. If `includeAuditLogs=true` → fetch from `audit_logs`.
4. **Synchronous** if `len(events) < 10`: generate PDF inline, return bytes.
5. **Asynchronous** if `len(events) >= 10`:
   - Store job status `PROCESSING` in Redis (TTL 24h, key: `trace_export_job:{jobID}`).
   - Launch goroutine: generate PDF → write to `storage/temp/` → update Redis to `COMPLETED` with `downloadUrl` → publish `trace.exported` event → write audit log.
   - Return `202 Accepted` with `ExportJobResponse`.

### Export Excel Flow
1. Either `productItemId` or `batchId` required.
2. Fetch product items (single or by batch).
3. For each item, fetch events (with optional date filter).
4. Guard: `totalEvents > 50000` → reject.
5. **Synchronous** if single item and `totalEvents < 10`.
6. Otherwise async (same Redis job pattern as PDF).
7. Excel has 2 sheets: "General Info" (item summary) + "Timeline Log" (all events flat).

### Repository (`TraceRepository`)
| Method | Query Description |
|---|---|
| `FindProductItemByCode` | product_items JOIN product_variants JOIN products — match by item_code OR serial_number |
| `FindProductItemsByBatchID` | product_items by batch_id |
| `FindEvents` | events LEFT JOIN locations LEFT JOIN users — filter by item_id, batch_id, date range, event types |
| `FindAuditLogs` | audit_logs LEFT JOIN users — filter by ProductItem entity or Batch entity |

### Valid Event Types (whitelist in code)
`PRODUCED, PACKED, WAREHOUSE_IN, WAREHOUSE_OUT, TRANSPORTED, SALE, REGISTERED, WARRANTY_ACTIVE, WARRANTY_CLAIM, WARRANTY_RESOLVED, RETURNED, RECALL, RECALLED`

### Database Tables Used
`product_items`, `product_variants`, `products`, `events`, `locations`, `users`, `audit_logs`

### Dependencies
- Redis (job state for async export, rate limiter)
- RabbitMQ publisher (`trace.exported`)
- `audit_log` pkg
- `storage/temp/` filesystem (PDF/Excel files for async jobs)

---

## 4. Module Event

### Purpose
Ghi nhận các sự kiện xảy ra trên một `product_item` trong vòng đời truy xuất nguồn gốc.

### Event Lifecycle
- Events được tạo qua `EventService.CreateEvent`.
- Không có endpoint HTTP trực tiếp registered trong router hiện tại (module `event` **không được wire vào router**).
- `EventRepository` sử dụng raw SQL (`database/sql`) thay vì GORM.
- Events được đọc lại bởi module `Trace` và `Batch` thông qua GORM queries trực tiếp vào bảng `events`.

### Entity (`Event`)
| Field | Type | Description |
|---|---|---|
| `id` | UUID PK | |
| `product_item_id` | UUID FK | Liên kết với product_items |
| `event_type` | varchar(100) | PRODUCED, PACKED, SHIPPED, SOLD, SCANNED... |
| `occurred_at` | timestamp | Thời điểm sự kiện xảy ra |
| `location` | varchar(255) | Free-text location |
| `actor` | varchar(255) | Người thực hiện |
| `description` | text | |
| `metadata` | jsonb | Arbitrary JSON payload |
| `is_deleted` | bool | Soft-delete flag |

> **Note**: Entity `event.go` dùng field `occurred_at`, `location`, `actor` (free-text). Tuy nhiên, bảng DB thực tế mà Trace query có các cột `title`, `description`, `location_id` (FK), `actor_id` (FK). Có sự khác biệt giữa schema entity Go và schema bảng thực tế được query trong `trace_repository.go`.

### Request DTO
`CreateEventRequest`: `product_item_id`, `event_type`, `occurred_at?`, `location`, `actor`, `description`, `metadata`.

### Business Rules
- `occurred_at` defaults to `time.Now()` if not provided.
- Module không có handler riêng — không expose HTTP endpoint nào.
- Events được đọc trong Trace search: filtered by role, date range, event type whitelist.
- Events được kiểm tra tồn tại trong `Batch.DeleteBatch` (`ExistsEvents`) để block xóa.

### Quan hệ với Trace và Batch
- **Trace** đọc events từ `events` table JOIN `locations`, `users`.
- **Batch** kiểm tra events tồn tại qua `product_items.batch_id → events.product_item_id`.
- **Batch.GetBatchEvents** query events qua `product_items` JOIN `events` filtered by `batch_id`.

---

## 5. Database Summary

### batches
- **Chức năng**: Lô hàng sản xuất/nhập khẩu
- **PK**: `id` (UUID)
- **Key fields**: `batch_code` (unique), `variant_id` (FK), `status`, `quantity`, `manufacture_date`, `expiry_date`, `origin_country`, `is_deleted`, `created_by`
- **Relations**: belongs to `product_variants`; has many `product_items`

### product_items
- **Chức năng**: Sản phẩm đơn lẻ trong lô, có QR code
- **PK**: `id` (UUID)
- **Key fields**: `item_code` (unique), `serial_number`, `verification_token`, `status`, `variant_id`, `batch_id`, `current_location_id`, `is_deleted`
- **Relations**: belongs to `batches`, `product_variants`; has many `events`; optional FK to `locations`

### events
- **Chức năng**: Timeline sự kiện vòng đời sản phẩm
- **PK**: `id` (UUID)
- **Key fields**: `product_item_id`, `batch_id` (nullable), `event_type`, `title`, `description`, `location_id` (FK), `actor_id` (FK), `created_at`, `is_deleted`
- **Relations**: belongs to `product_items`; optional FK to `locations`, `users`

> ⚠️ Schema bảng `events` có `location_id` (FK) và `actor_id`/`actor` — cần cross-check migration thực tế. Entity Go dùng free-text `location` và `actor`.

### audit_logs
- **Chức năng**: Lịch sử thay đổi hệ thống (ai làm gì với entity nào)
- **PK**: `id` (UUID)
- **Key fields**: `user_id` (nullable FK), `action` (CREATE/UPDATE/DELETE), `entity`, `entity_id`, `old_data` (jsonb), `new_data` (jsonb), `created_at`
- **Relations**: optional FK to `users`

### products
- **Chức năng**: Sản phẩm (danh mục)
- **PK**: `id` (UUID)
- **Key fields**: `name`, `thumbnail_url`, `category_id`
- **Relations**: has many `product_variants`

### product_variants
- **Chức năng**: Biến thể sản phẩm (màu, size...)
- **PK**: `id` (UUID)
- **Key fields**: `product_id` (FK), `name`
- **Relations**: belongs to `products`; has many `batches`, `product_items`

### locations
- **Chức năng**: Kho, địa điểm lưu trữ
- **PK**: `id` (UUID)
- **Key fields**: `name`, `type`
- **Relations**: referenced by `product_items.current_location_id`, `events.location_id`

### users
- **Chức năng**: Tài khoản người dùng
- **PK**: `id` (UUID)
- **Key fields**: `email`, `full_name`, `role`, `is_deleted`
- **Relations**: referenced by `audit_logs.user_id`, `batches.created_by`, `events.actor_id`

---

## 6. Request Flow

### Standard HTTP Request
```
HTTP Request
  ↓
Gin Engine
  ↓
Global Middleware (Recovery → RequestID → Logger)
  ↓
Route-level Middleware (AuthMiddleware → RoleMiddleware)
  ↓
Handler (parse params, bind DTO, call service)
  ↓
Service (business logic, validation, orchestration)
  ↓
Repository (GORM queries to PostgreSQL)
  ↓
Database (PostgreSQL)
```

### With Message Queue
```
Service
  ↓ (non-blocking, fire-and-forget)
publisher.Publish(Event)
  ↓
RabbitMQ Exchange: product-trace.events
  ↓ (routing key)
Queue → Consumer (other services or notification service)
```

### Async Export Flow (Trace PDF/Excel with >= 10 events)
```
POST /api/trace/export/pdf
  ↓
Handler → Service
  ↓ (immediate)
Redis.Set(jobKey, PROCESSING, 24h)
  ↓ (goroutine, async)
Generate PDF → Write storage/temp/ → Redis.Set(COMPLETED) → Publish trace.exported → AuditLog
  ↓ (immediate response)
202 Accepted { jobId, status: PROCESSING }
```

---

## 7. Important Business Rules

### Batch
- `batch_code` auto-generated: `{PREFIX}-{YEAR}-{NNNN}`. Prefix normalized to uppercase.
- Advisory lock (`pg_advisory_xact_lock`) per `prefix-year` prevents concurrent duplicate sequence.
- `CreateBatch` retries 3 times (100ms backoff) for transient DB errors.
- Status enum: `DRAFT, ACTIVE, EXPIRED, RECALLED, LOCKED` (filter); `ACTIVE, EXPIRED, RECALLED, BLOCKED` (update).
- Only ADMIN can see/filter DRAFT batches. Non-ADMIN with `status=DRAFT` → 403.
- Soft-delete blocked if batch has any linked product items OR events.
- Export decrements `batch.quantity` in a transaction, records in `audit_logs`.
- `audit_log` write failures are silently ignored (log error, do not block main response).

### Trace
- Search requires `code` 3–100 chars.
- Public/CUSTOMER role: `PACKED`, `WAREHOUSE_IN`, `WAREHOUSE_OUT`, `TRANSPORTED` events are stripped from results.
- Status `RECALLED` on product item triggers `WARNING` field in response.
- Async export: triggered when `len(events) >= 10` (PDF) or `len(events) >= 10 AND NOT single item` (Excel).
- Excel export hard limit: 50,000 total events — rejects with 400 if exceeded.
- PDF supports two themes: `CLASSIC_NAVY` (default), `WARM_MINIMAL`.
- PDF includes SHA-256 checksum watermark of item code + serial + event IDs.

### Event
- `EventService.CreateEvent` is the only write path. No HTTP endpoint is registered.
- `occurred_at` defaults to `time.Now()` if nil.

### Rate Limiting
- Trace search: 30 requests per minute per IP.
- Redis-backed; falls back to in-memory `sync.Mutex` map if Redis unavailable.

---

## 8. API Summary

| Method | Endpoint | Module | Mô tả |
|---|---|---|---|
| POST | `/api/auth/register` | Auth | Đăng ký tài khoản |
| POST | `/api/auth/login` | Auth | Đăng nhập |
| POST | `/api/auth/verify-otp` | Auth | Xác thực OTP |
| POST | `/api/auth/resend-otp` | Auth | Gửi lại OTP |
| POST | `/api/auth/refresh` | Auth | Refresh token |
| POST | `/api/auth/logout` | Auth | Đăng xuất |
| POST | `/api/auth/forgot-password` | Auth | Quên mật khẩu |
| POST | `/api/auth/reset-password` | Auth | Đặt lại mật khẩu |
| GET | `/api/users/profile` | User | Xem profile cá nhân |
| PUT | `/api/users/profile/:id` | User | Cập nhật profile |
| POST | `/api/users` | User | Tạo user (ADMIN) |
| PUT | `/api/users/:id` | User | Cập nhật user (ADMIN) |
| DELETE | `/api/users/:id` | User | Xóa user (ADMIN) |
| GET | `/api/users` | User | Danh sách users (ADMIN) |
| GET | `/api/users/:id` | User | Chi tiết user (ADMIN) |
| GET | `/api/batches/:id` | Batch | Chi tiết lô (public, nhận batch_code hoặc UUID) |
| GET | `/api/batches` | Batch | Danh sách lô (JWT) |
| GET | `/api/batches/search` | Batch | Tìm kiếm lô (JWT) |
| GET | `/api/batches/:id/events` | Batch | Events của lô (JWT) |
| GET | `/api/batches/:id/products` | Batch | Sản phẩm trong lô (JWT) |
| GET | `/api/batches/:id/history` | Batch | Lịch sử thay đổi (ADMIN, STAFF) |
| GET | `/api/batches/export-qr/:id` | Batch | Xuất PDF QR code (ADMIN, MANUFACTURER) |
| POST | `/api/batches` | Batch | Tạo lô (ADMIN, MANUFACTURER) |
| PATCH | `/api/batches/:id/status` | Batch | Cập nhật trạng thái (ADMIN, MANUFACTURER) |
| DELETE | `/api/batches/:id` | Batch | Xóa mềm lô (ADMIN, MANUFACTURER) |
| POST | `/api/batches/:id/export` | Batch | Xuất hàng khỏi lô (ADMIN, MANAGER, WAREHOUSE) |
| GET | `/api/trace/search` | Trace | Tra cứu timeline sản phẩm (public + rate limit) |
| POST | `/api/trace/export/pdf` | Trace | Xuất PDF timeline (JWT) |
| POST | `/api/trace/export/excel` | Trace | Xuất Excel timeline (JWT) |
| GET | `/api/products` | Product | Danh sách sản phẩm (public) |
| GET | `/api/products/:id` | Product | Chi tiết sản phẩm (public) |
| POST | `/api/products` | Product | Tạo sản phẩm (ADMIN, MANUFACTURER) |
| PUT | `/api/products/:id` | Product | Cập nhật sản phẩm (ADMIN, MANUFACTURER) |
| DELETE | `/api/products/:id` | Product | Xóa sản phẩm (ADMIN) |
| GET | `/api/variants/:id` | Variant | Chi tiết variant (public) |
| GET | `/api/variants/product/:product_id` | Variant | Variants theo product (public) |
| PUT | `/api/variants/:id` | Variant | Cập nhật variant (ADMIN, MANUFACTURER) |
| DELETE | `/api/variants/:id` | Variant | Xóa variant (ADMIN) |
| GET | `/api/variants/:id/attributes` | Variant | Attributes của variant (public) |
| POST | `/api/variants/:id/attributes` | Variant | Gán attributes (ADMIN, MANUFACTURER) |
| GET | `/api/attributes` | Attribute | Danh sách attributes (public) |
| GET | `/api/attributes/:id` | Attribute | Chi tiết attribute (public) |
| POST | `/api/attributes` | Attribute | Tạo attribute (ADMIN, MANUFACTURER) |
| PUT | `/api/attributes/:id` | Attribute | Cập nhật attribute (ADMIN, MANUFACTURER) |
| DELETE | `/api/attributes/:id` | Attribute | Xóa attribute (ADMIN) |
| GET | `/api/attribute-values` | AttrValue | Danh sách values (public) |
| GET | `/api/attribute-values/:id` | AttrValue | Chi tiết value (public) |
| PUT | `/api/attribute-values/:id` | AttrValue | Cập nhật value (ADMIN, MANUFACTURER) |
| DELETE | `/api/attribute-values/:id` | AttrValue | Xóa value (ADMIN) |
| GET | `/api/locations` | Location | Danh sách locations (public) |
| GET | `/api/locations/:id` | Location | Chi tiết location (public) |
| POST | `/api/locations` | Location | Tạo location (ADMIN) |
| PUT | `/api/locations/:id` | Location | Cập nhật location (ADMIN) |
| DELETE | `/api/locations/:id` | Location | Xóa location (ADMIN) |
| GET | `/api/public/verify` | Public | Xác thực QR (anonymous) |

---

## 9. Module Dependencies

```
Batch
 ├── product_variants (FK check on create)
 ├── product_items (QR export, constraint checks)
 ├── events (constraint check on delete, GetBatchEvents)
 ├── locations (GetBatchProducts: current_location_id)
 ├── audit_logs (update/delete/history/export)
 └── RabbitMQ (batch.created, batch.history_viewed)

Trace
 ├── product_items (lookup by code)
 ├── product_variants → products (product name, thumbnail)
 ├── events (timeline)
 ├── locations (event location)
 ├── users (event actor, audit user email)
 ├── audit_logs (optional appendix in PDF export)
 ├── Redis (async job state, rate limiting)
 ├── RabbitMQ (trace.exported)
 └── filesystem: storage/temp/ (async export files)

Event (write)
 └── product_items (product_item_id FK)

Auth
 ├── users
 ├── Redis (OTP storage via cache pkg)
 └── RabbitMQ (otp.registered, otp.forgot, user.registered)

User
 └── RabbitMQ (user events)

Public
 └── product_items (verification_token lookup)
```

---

## 10. Current Limitations

1. **Module Event không có HTTP endpoint**: `EventService` và `EventRepository` tồn tại nhưng không được wire vào router. Không có cách tạo event qua API.

2. **Schema mismatch Event**: Entity Go (`event.go`) dùng `location` (free-text), `actor` (free-text). Trace repository query bảng `events` với `location_id` (FK to locations) và `actor_id` (FK to users). Hai schema không khớp.

3. **Worker không implement**: `cmd/worker/` entrypoint và `internal/workers/` tồn tại nhưng không có logic. RabbitMQ consumer (`consumer.go`) được implement nhưng không được khởi động trong bất kỳ entrypoint nào.

4. **Async export không có status check endpoint**: Job status được lưu trong Redis (`trace_export_job:{jobID}`) nhưng không có API `GET /trace/export/status/:jobId` để client poll.

5. **`GetBatchDetail` nhận cả UUID và batch_code**: Handler dùng `c.Param("id")` làm string và gọi `FindByCode`. Nếu client truyền UUID, sẽ không tìm thấy trừ khi batch_code trùng với UUID string.

6. **`product_item` router bị comment out**: `SetupProductItemRouter` bị comment trong `router.go`. Module `product_item` có handler nhưng không có route.

7. **Audit log write failure bị silent**: Lỗi ghi audit log không được return hay alert — chỉ được `_ = logErr`.

8. **`realtime/` package tồn tại nhưng không được sử dụng** trong bất kỳ flow nào được wire vào app.
