# Event-Driven Architecture Review

Ngày đánh giá: 2026-07-01  
Phạm vi: Chỉ dựa trên phần code hiện có, bỏ qua TODO và module chưa hoàn thiện.

## 1) Service/Module nên phát sự kiện (Publisher)

### Auth Service (go-core-service)
- Hành động nên publish:
  - RegisterUser
  - VerifyOTP
  - LoginUser
  - Logout
- Event name đề xuất:
  - user.registered (đã có)
  - user.verified (đã có)
  - user.logged_in (đã có hằng số routing key)
  - user.logged_out (đề xuất)
- Payload nên chứa:
  - eventId, eventType, eventVersion, timestamp, producer, correlationId
  - userId, email, role, occurredAt
  - riêng user.registered có thể chứa full_name; cần hạn chế phát tán otp_code
- Lý do publish:
  - Tách side-effect (email, audit, analytics, security) khỏi transaction chính.

### User Service (go-core-service)
- Hành động nên publish:
  - UpdateProfile (đã có)
  - CreateUser, UpdateUser, DeleteUser
- Event name đề xuất:
  - user.profile_updated (đã có)
  - user.created
  - user.updated
  - user.deleted
- Payload nên chứa:
  - userId, actorId, changedFields, updatedAt
- Lý do publish:
  - Đồng bộ read model/search index và audit tập trung.

### Product Service (go-core-service)
- Hành động nên publish:
  - CreateProduct
  - UpdateProduct
  - DeleteProduct
- Event name đề xuất:
  - product.created
  - product.updated
  - product.deleted
- Payload nên chứa:
  - productId, categoryId, name, status, variantSummary, changedAt
- Lý do publish:
  - Kích hoạt các tác vụ async như AI indexing/search.

### Batch Service (go-core-service)
- Hành động nên publish:
  - CreateBatch
- Event name đề xuất:
  - batch.created
- Payload nên chứa:
  - batchId, batchCode, variantId, quantity, manufactureDate, expiryDate
- Lý do publish:
  - Cập nhật trace projection và analytics downstream.

### Location Service (go-core-service)
- Hành động nên publish:
  - CreateLocation
  - UpdateLocation
  - HardDeleteLocation
- Event name đề xuất:
  - location.created
  - location.updated
  - location.deleted
- Payload nên chứa:
  - locationId, ownerUserId, code, type, city, geo
- Lý do publish:
  - Phù hợp cho geo-search, cache invalidation, đồng bộ dữ liệu đọc.

### Event Service (go-core-service/internal/modules/event)
- Hành động nên publish:
  - CreateEvent
- Event name đề xuất:
  - trace.created
- Payload nên chứa:
  - productItemId, eventType, locationId, occurredAt, metadata
- Lý do publish:
  - Là luồng domain event cốt lõi cho traceability.

---

## 2) Service/Module nên nhận sự kiện (Consumer)

### user.registered
- Consumer:
  - Notification worker (nest-ai-service)
  - Audit worker (đề xuất)
- Mục đích consume:
  - Gửi welcome/OTP email, ghi audit.
- Retry: Có
- DLQ: Có

### user.verified
- Consumer:
  - Notification worker
  - Analytics/Loyalty worker (đề xuất)
- Mục đích consume:
  - Gửi xác nhận, cập nhật activation metrics.
- Retry: Có
- DLQ: Có

### product.created
- Consumer:
  - AI indexing/search worker
- Mục đích consume:
  - Tạo embedding, index dữ liệu.
- Retry: Có
- DLQ: Có

### batch.created
- Consumer:
  - Trace projection worker
  - Analytics worker
- Mục đích consume:
  - Xây read model và báo cáo.
- Retry: Có
- DLQ: Có

### user.profile_updated
- Consumer:
  - Search/profile index worker
  - Audit worker
- Mục đích consume:
  - Đồng bộ dữ liệu đọc, lưu lịch sử thay đổi.
- Retry: Có
- DLQ: Có

Ghi chú quan trọng:
- Hiện có mismatch routing key password reset giữa hai service:
  - go-core-service: user.password_reset_requested
  - nest-ai-service: auth.password_reset_requested

---

## 3) Phân tích dependency giữa các service

### Service nào đang gọi trực tiếp service khác
- Chưa thấy call HTTP/gRPC trực tiếp giữa go-core-service và nest-ai-service trong phần code nghiệp vụ hiện có.
- Mức vào hệ thống đi qua Kong Gateway và route riêng:
  - /api -> go-core-service
  - /ai -> nest-ai-service

### Chỗ nên chuyển sang Event-Driven
- Các tác vụ có side-effect không cần phản hồi ngay:
  - product.created, product.updated, product.deleted
  - batch.created
  - location.created, location.updated, location.deleted
  - user.created, user.updated, user.deleted, user.profile_updated
  - trace.created
- Lý do:
  - Giảm coupling, tăng khả năng scale độc lập theo queue.

### Chỗ vẫn nên synchronous REST/gRPC
- API cần phản hồi tức thì cho client:
  - auth register/login/verify
  - CRUD core nghiệp vụ
- Lý do:
  - Cần tính nhất quán và phản hồi đồng bộ ngay lập tức.

---

## 4) Thiết kế RabbitMQ đề xuất

### Exchange
- product-trace.events (topic)
- product-trace.retry (topic)
- product-trace.dlx (topic/direct)

### Queue
- notification.user.registered.q
- notification.user.verified.q
- ai.product.created.q
- ai.batch.created.q
- audit.user.events.q
- ai.events.dlq (hoặc tách domain-specific dlq)

### Routing Key
- user.registered
- user.verified
- user.profile_updated
- product.created
- product.updated
- batch.created
- trace.created
- *.failed

### Binding mẫu
- notification.user.registered.q <- product-trace.events : user.registered
- ai.product.created.q <- product-trace.events : product.created
- audit.user.events.q <- product-trace.events : user.*
- retry queues <- product-trace.retry : cùng routing key nghiệp vụ
- dlq queues <- product-trace.dlx : *.failed

---

## 5) Thiết kế Worker (phần trọng tâm)

## Phương án 1: Một Worker consume một Queue

Ví dụ:
- Worker Notification -> notification.user.registered.q
- Worker Audit -> audit.user.events.q
- Worker AI -> ai.product.created.q

Ưu điểm:
- Cô lập lỗi tốt.
- Scale độc lập theo từng nghiệp vụ.
- Dễ tinh chỉnh prefetch/concurrency riêng.

Nhược điểm:
- Tăng số process/deployment.
- Tăng chi phí vận hành và quan sát.

Khi nào nên dùng:
- Traffic lớn, SLA khác nhau theo luồng.
- Workload khác biệt rõ giữa các queue.

Khả năng mở rộng:
- Rất tốt theo chiều ngang từng queue.

Khả năng scale:
- Linh hoạt nhất, tránh ảnh hưởng chéo.

---

## Phương án 2: Một Worker consume nhiều Queue hoặc nhiều EventPattern

Ví dụ:
- Notification Worker
  - notification.user.registered.q
  - notification.user.verified.q
  - notification.password.reset.q

Ưu điểm:
- Triển khai nhanh, vận hành đơn giản.
- Ít service/process hơn.
- Phù hợp giai đoạn đang phát triển.

Nhược điểm:
- Cô lập lỗi kém hơn phương án 1.
- Khó tune tối ưu theo từng event type.
- Một worker quá tải có thể ảnh hưởng luồng khác.

Khi nào nên dùng:
- MVP/early stage.
- Lưu lượng chưa cao.

Khả năng scale:
- Scale được, nhưng kém linh hoạt hơn phương án 1.

---

## Kết luận mô hình Worker phù hợp nhất cho codebase hiện tại

Nên chọn Phương án 2 ở thời điểm hiện tại.

Lý do:
- Hệ thống hiện đang theo mô hình queue chung ai.events và nhiều consumer trong cùng app nest-ai-service.
- Chưa có worker process tách riêng trong codebase hiện hữu.
- Phù hợp với giai đoạn đang phát triển, giảm độ phức tạp vận hành.

Lộ trình nâng cấp:
- Khi product.created, batch.created, trace.created tăng tải hoặc SLA tách biệt, chuyển dần sang Phương án 1.

---

## 6) Khả năng Scale

- Có nên tách worker thành nhiều process không: Có, theo lộ trình từng bước.
- Có nên chạy nhiều instance cùng worker không: Có.
- RabbitMQ phân phối message như thế nào:
  - Cùng queue, nhiều consumer: competing consumers (chia tải).
  - Nhiều queue cùng bind key: fan-out theo binding.
- Có cần prefetch count không: Có.
  - Gợi ý khởi điểm:
    - notification: 20-50
    - ai/indexing: 5-10
- Có cần idempotency không: Có, bắt buộc trong at-least-once delivery.
- Có cần retry queue và DLQ không: Có.
  - Retry cho lỗi tạm thời.
  - DLQ cho lỗi quá ngưỡng retry.

---

## 7) Sơ đồ kiến trúc đề xuất (ASCII)

Go Core Service (Auth/Product/Batch/User/Location)
        |
        | publish domain events
        v
product-trace.events (topic exchange)
        |
        |-------------------------------|
        |                               |
        v                               v
notification.user.q               ai.product.q
        |                               |
        v                               v
Notification Worker               AI Index Worker
        |                               |
        v                               v
SendGrid/Email                     Search/Vector Projection

        |
        | (fail after retries)
        v
product-trace.dlx
        |
        v
*.dlq queues (manual replay)

---

## Bảng tổng kết

| Service | Publish Event | Consume Event | Queue | Worker | Ghi chú |
|---|---|---|---|---|---|
| go-core-service/auth | user.registered, user.verified, user.logged_in, user.logged_out | Không | product-trace.events (exchange) | Publisher trong API process | Đã publish registered/verified |
| go-core-service/user | user.profile_updated, user.created, user.updated, user.deleted | Không | product-trace.events | Publisher trong API process | profile_updated đã có |
| go-core-service/product | product.created, product.updated, product.deleted | Không | product-trace.events | Publisher trong API process | Có CRUD, nên phát event |
| go-core-service/batch | batch.created | Không | product-trace.events | Publisher trong API process | CreateBatch đã có |
| go-core-service/location | location.created, location.updated, location.deleted | Không | product-trace.events | Publisher trong API process | Có create/update/delete |
| go-core-service/event | trace.created | Không | product-trace.events | Publisher trong API process | Có CreateEvent domain |
| nest-ai-service/notification | Không | user.registered, user.verified, password_reset_requested | ai.events (hiện tại), notification.*.q (đề xuất) | Hiện: worker chung; tương lai: tách dần | Manual ack/nack đã có |
| nest-ai-service/ai-indexing | Không | product.created, batch.created, trace.created | ai.events (hiện tại), ai.*.q (đề xuất) | Hiện: worker chung; tương lai: tách khi tải tăng | Phù hợp xử lý async nặng |
| nest-ai-service/audit | Không | user.*, product.*, batch.* | audit.*.q | Audit worker | Nên tách độc lập để scale |

## Kết luận cuối

Với codebase hiện tại, mô hình worker phù hợp nhất là mô hình worker dùng chung (consume nhiều event/queue) để tối ưu tính đơn giản và tốc độ triển khai. Khi tải tăng và SLA phân hóa, nên tách dần thành worker chuyên biệt theo domain để đạt khả năng scale và cô lập lỗi tốt hơn.
