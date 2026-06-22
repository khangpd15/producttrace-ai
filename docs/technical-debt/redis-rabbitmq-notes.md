# Technical Debt: Redis & RabbitMQ Scaling Notes

Tài liệu này ghi nhận các vấn đề thiết kế hệ thống phân tán liên quan đến Redis và RabbitMQ trong Go Core Service có thể ảnh hưởng đến hiệu năng và tính ổn định khi hệ thống scale lên quy mô lớn.

---

## 1. Sử dụng SCAN để tìm kiếm và xóa khóa (Redis DeletePattern)

* **Mô tả**: Hàm `DeletePattern` trong `RedisCache` dùng lệnh `SCAN` để duyệt qua toàn bộ keyspace nhằm tìm các khóa khớp với mẫu pattern (ví dụ: `refresh_token:*:<token_hash>`) để xóa chúng.
* **Mức độ ưu tiên**: Trung bình (Medium)
* **Tác động**: Khi số lượng active sessions (hoặc số lượng keys trong Redis) tăng lên hàng triệu, lệnh `SCAN` sẽ quét qua toàn bộ DB, gây tăng cao mức sử dụng CPU của Redis và làm tăng đáng kể độ trễ (latency) của API Refresh Token.
* **Khi nào cần xử lý**: Khi số lượng key trong Redis vượt quá 50,000 keys hoặc latency của API refresh tăng lên trên 100ms.
* **Giải pháp đề xuất**:
  * Tránh dùng `SCAN`/`KEYS` hoàn toàn.
  * Sử dụng cấu trúc dữ liệu **Redis Set** để quản lý quan hệ 1-N. Ví dụ, lưu danh sách token hash của một user dưới key Set `user:tokens:<user_id>`.
  * Khi tạo token: `SADD user:tokens:<user_id> <token_hash>`.
  * Khi thu hồi toàn bộ token của user: Chỉ cần lấy danh sách từ Set bằng `SMEMBERS` (hoặc xóa trực tiếp bằng Set operations) rồi xóa các key cụ thể trong $O(1)$.
* **Trade-off**: Tăng thêm dung lượng lưu trữ nhẹ trong Redis (để lưu Set) và tăng thêm độ phức tạp trong logic code khi phải quản lý đồng thời cả key string và Set key.

---

## 2. Lỗi Dual-Write & Nguy cơ mất mát sự kiện (Event Loss)

* **Mô tả**: Tận dụng trực tiếp publisher của RabbitMQ để phát sự kiện đăng ký (`user.registered`) ngay trong luồng HTTP request.
* **Mức độ ưu tiên**: Cao (High)
* **Tác động**: Nếu PostgreSQL commit lưu thông tin User thành công, nhưng ngay sau đó mạng bị chập chờn hoặc RabbitMQ bị quá tải làm lệnh `Publish` thất bại, người dùng sẽ không nhận được OTP kích hoạt tài khoản. Điều này phá vỡ tính nhất quán cuối cùng (Eventual Consistency).
* **Khi nào cần xử lý**: Ngay trước khi đưa hệ thống chạy thử nghiệm Beta hoặc khi tỷ lệ mất mát sự kiện được phát hiện > 0.01%.
* **Giải pháp đề xuất**:
  * Áp dụng **Transactional Outbox Pattern**.
  * Tạo bảng `outbox_events` trong PostgreSQL.
  * Khi đăng ký User, lưu bản ghi User và bản ghi Outbox event vào cùng một Database Transaction.
  * Thiết lập một background worker (CDC/Debezium hoặc polling worker bằng Go) quét bảng `outbox_events` định kỳ, publish lên RabbitMQ và đánh dấu sự kiện đã gửi thành công khi nhận được Publisher Confirm.
* **Trade-off**: Tốn thêm tài nguyên PostgreSQL để ghi/xóa dữ liệu bảng outbox liên tục và tăng độ trễ gửi tin nhắn đi một vài mili-giây (do chạy bất đồng bộ qua worker).

---

## 3. Khả năng chống chịu lỗi của Cache (Cache Resilience / Fail-silent)

* **Mô tả**: Module `RedisCache` ném trực tiếp lỗi kết nối lên tầng Service khi Redis bị sập. Ở tầng Service, các lỗi này khiến API trả về HTTP 500.
* **Mức độ ưu tiên**: Thấp (Low)
* **Tác động**: Với các tính năng bắt buộc cần Redis như OTP và Session (Refresh Token), việc fail-fast là chính xác. Tuy nhiên, đối với các dữ liệu cache thuần túy (ví dụ: cache danh sách batch sản phẩm), việc sập Redis không nên làm hỏng API mà nên "fail-silent" và đọc trực tiếp từ Database.
* **Khi nào cần xử lý**: Khi bắt đầu áp dụng Redis làm cache lớp đọc (Read Cache) cho các API truy vấn sản phẩm/lô hàng.
* **Giải pháp đề xuất**:
  * Thiết lập cơ chế fallback trong Cache layer: Nếu Redis báo lỗi kết nối, tự động log lỗi lại và trả về lỗi giả (`ErrCacheMiss`) để service tự động nhảy vào database đọc dữ liệu thay vì trả lỗi về client.
* **Trade-off**: Database có thể bị quá tải (Cache Stampede) nếu Redis sập đột ngột trong thời gian lưu lượng truy cập cao.

---

## 4. Thiếu Tracing Context Propagation qua RabbitMQ

* **Mô tả**: Tin nhắn đẩy lên RabbitMQ hiện chưa đính kèm thông tin tracing headers (như W3C `traceparent` hoặc Uber `uber-trace-id`).
* **Mức độ ưu tiên**: Trung bình (Medium)
* **Tác động**: Không thể theo dõi được vết của một request đi từ Go Core Service (khi phát sự kiện) sang NestJS AI/Worker Service (khi tiêu thụ sự kiện). Điều này gây cực kỳ khó khăn cho việc debug trong môi trường microservices.
* **Khi nào cần xử lý**: Khi hệ thống có trên 3 microservices giao tiếp qua lại và bắt đầu triển khai các giải pháp giám sát như OpenTelemetry + Jaeger/Grafana Tempo.
* **Giải pháp đề xuất**:
  * Trích xuất OpenTelemetry Context từ `context.Context` trong Go và chèn vào trường Headers của tin nhắn RabbitMQ (AMQP Properties Headers) khi publish. Phía NestJS sẽ giải mã headers này để tiếp tục span tracing.
* **Trade-off**: Tăng kích thước tin nhắn và yêu cầu cài đặt SDK OpenTelemetry ở cả 2 service.

---

## 5. Đảm bảo tính Idempotency ở phía Consumer

* **Mô tả**: RabbitMQ đảm bảo cơ chế phân phối tin nhắn "at-least-once" (ít nhất một lần). Trong trường hợp mạng chập chờn, consumer có thể nhận trùng lặp tin nhắn (duplicate events).
* **Mức độ ưu tiên**: Cao (High)
* **Tác động**: Nếu một sự kiện như `user.verified` bị xử lý trùng lặp, có thể dẫn đến việc thực thi lại các tác vụ phụ trợ (như gửi email chúc mừng, kích hoạt ưu đãi) nhiều lần.
* **Khi nào cần xử lý**: Khi xây dựng các worker xử lý thanh toán, kích hoạt tài khoản hoặc trừ điểm tích lũy.
* **Giải pháp đề xuất**:
  * Phía Consumer (bất kể Go hay NestJS) cần có cơ chế kiểm tra trùng lặp bằng cách kiểm tra `EventID` trong cơ sở dữ liệu trước khi xử lý, hoặc lưu trạng thái các sự kiện đã xử lý vào Redis với TTL ngắn hạn.
* **Trade-off**: Yêu cầu kiểm tra database hoặc cache trước mỗi lần xử lý tin nhắn, làm tăng nhẹ độ trễ xử lý sự kiện.
