# Báo Cáo Nhận Xét & Đánh Giá Hệ Thống Hiện Tại - ProductTrace AI

Tài liệu này tổng hợp các nhận xét, đánh giá kỹ thuật đối với hệ thống hiện tại của dự án **ProductTrace AI** sau khi hoàn thành đợt audit mã nguồn và refactor các middleware cốt lõi.

---

## 1. Đánh Giá Kiến Trúc & Cấu Trúc Dự Án (Architecture & Structure)

### Điểm Tốt (Strengths)
* **Phân tách Layer rõ ràng**: Dự án tuân thủ tốt triết lý Clean Architecture. Các module nghiệp vụ (`user`, `authen`, `batch`, `product`) có cấu trúc đồng nhất, tách biệt rõ ràng giữa Entities, Repositories, Services và Handlers.
* **Sử dụng Dependency Injection (DI)**: Các dependencies được khai báo thông qua Interfaces và inject thông qua các hàm Constructor (ví dụ: `NewUserRepository`, `NewbatchService`). Việc này giúp code dễ test, dễ mock và giảm thiểu coupling.
* **Giao tiếp hướng sự kiện (Event-Driven)**: Hệ thống sử dụng RabbitMQ làm cầu nối bất đồng bộ giữa Go Core và Nest AI. Việc cấu hình Connection Manager có cơ chế tự động reconnect và Publisher Confirm giúp luồng gửi message tin cậy hơn.

### Điểm Yếu & Rủi Ro Cấu Trúc (Weaknesses & Risks)
* **Duy trì song song hai Connection Pool**: Go Core Service khởi tạo hai pool kết nối độc lập tới cùng một PostgreSQL (một cho GORM và một cho raw SQL thông qua `sql.DB` của thư viện `database/sql`). Việc này gây lãng phí tài nguyên RAM/Connection của Postgres và khiến các thao tác ghi của hai tầng không thể chạy chung trong một Database Transaction duy nhất.
* **Code Skeleton chưa hoàn thiện**: Một số module trong dự án hiện tại chỉ đang ở dạng skeleton (ví dụ: `location`, `product_category` trong Go chưa được cấu hình định tuyến; Nest AI Service chưa có các logic nghiệp vụ thực tế ngoài consumer).

---

## 2. Nhận Xét Chi Tiết Các Lỗi Đã Được Khắc Phục (Refactored Middlewares)

Trong đợt refactor vừa qua, các lỗi nghiêm trọng ảnh hưởng trực tiếp tới tính an toàn và hiệu năng của hệ thống đã được xử lý triệt để:

### 2.1. Middleware Xác Thực (`AuthMiddleware`)
* **Trạng thái cũ**:
  * Import sai đường dẫn package nghiệp vụ (`github.com/product-trace-ai/...`).
  * Có lỗi cú pháp gãy code ở cuối file dẫn đến việc chương trình không thể biên dịch.
  * Truy vấn database trên **mọi request** (`userRepo.GetUserByID`) để xác thực, làm mất đi lợi thế stateless của JWT và tăng tải cho DB.
  * Chỉ trích xuất được `user_id` mà bỏ sót `email` và `role`.
* **Trạng thái sau refactor**:
  * Sửa lỗi cú pháp và cập nhật import chính xác (`github.com/khangpd15/producttrace-ai/...`).
  * Trích xuất đầy đủ thông tin: `user_id`, `email`, `role` từ JWT Claims và lưu thẳng vào Gin Context.
  * **Bỏ qua query DB mặc định**, cải thiện hiệu năng xử lý request gấp nhiều lần.

### 2.2. Middleware Phân Quyền (`RoleMiddleware`)
* **Trạng thái cũ**:
  * Trả về lỗi `403 Forbidden` vô điều kiện đối với tất cả các request, khóa toàn bộ API có dùng middleware này.
  * Import sai package `"task_api/..."`.
* **Trạng thái sau refactor**:
  * Hỗ trợ kiểm tra quyền động đối với danh sách `allowedRoles ...string`.
  * Nhận diện role trực tiếp từ context của AuthMiddleware, tự động fallback đọc từ `current_user` nếu có để đảm bảo tương thích ngược.

### 2.3. Middleware Phục Hồi Lỗi (`RecoveryMiddleware`)
* **Trạng thái cũ**:
  * Sử dụng ép kiểu cứng `err.(string)`. Nếu hệ thống bị panic bởi một đối tượng `error` hoặc struct bất kỳ, lệnh ép kiểu này sẽ gây ra lỗi crash lần thứ hai (panic within panic).
* **Trạng thái sau refactor**:
  * Sử dụng `switch err.(type)` để xử lý động mọi kiểu dữ liệu panic (string, error, struct) một cách an toàn.
  * Tự động ghi nhận thông tin panic ra log bằng `log.Printf` trước khi trả phản hồi HTTP 500 lỗi cho Client.

### 2.4. Middleware Định Danh Yêu Cầu (`RequestIDMiddleware`)
* **Trạng thái cũ**:
  * Sử dụng `UnixNano` làm Request ID.
* **Trạng thái sau refactor**:
  * Chuyển sang sử dụng **UUID v4** thông qua thư viện `github.com/google/uuid` đảm bảo tính duy nhất trên môi trường phân tán (nhiều bản sao container).

---

## 3. Các Lỗi Nghiêm Trọng Khác Cần Sửa (Infrastructure & Config Gaps)

Dưới đây là các điểm nghẽn nghiêm trọng được phát hiện tại cấu hình hạ tầng cần được đội ngũ phát triển khắc phục trước khi golive:

### 3.1. Lỗi Driver PostgreSQL của Go Core
* **Hiện trạng**: Hàm `ConnectPostgresSQL` sử dụng driver name `"pgx"` nhưng trong code chỉ import driver `"postgres"` (`_ "github.com/lib/pq"`).
* **Hậu quả**: Go Core Service sẽ bị crash lập tức khi khởi động với lỗi `unknown driver "pgx"`.
* **Khuyến nghị**: Đổi tên driver trong `sql.Open` thành `"postgres"`.

### 3.2. Lỗi Khớp Cổng Dịch Vụ Nest AI
* **Hiện trạng**: Nest AI Service chỉ chạy như một RMQ microservice (không mở HTTP port), nhưng Kong Gateway lại định tuyến HTTP `/ai` tới port 3000 của service này.
* **Hậu quả**: Các request gửi tới cổng API Gateway `/ai` sẽ bị lỗi `502 Bad Gateway` hoặc `Connection Refused`.
* **Khuyến nghị**: Chuyển Nest AI sang kiến trúc **Hybrid App** (vừa lắng nghe HTTP port 3000, vừa lắng nghe RabbitMQ).

### 3.3. Lỗi Đọc Biến Môi Trường `GetEnvAsInt`
* **Hiện trạng**: Logic hàm `GetEnvAsInt` trong `utils/helper.go` bị ngược điều kiện (`value != ""` trả về mặc định).
* **Hậu quả**: Các cấu hình dạng số (như `REDIS_PORT`, `POSTGRES_PORT`) luôn bị ghi đè về giá trị mặc định, bỏ qua cấu hình thực tế trong file `.env`.
* **Khuyến nghị**: Đổi điều kiện thành `value == ""`.

---

## 4. Đề Xuất Kế Hoạch Tiếp Theo (Action Plan)

1. **Áp dụng các Middleware đã Refactor vào Router**:
   * Tích hợp `AuthMiddleware` và `RoleMiddleware` vào [internal/router/router.go](file:///d:/producttrace-ai/apps/go-core-service/internal/router/router.go) để bảo vệ các API riêng tư (Profile, CRUD Users, Lô hàng, Sản phẩm) theo bảng phân loại phân quyền.
2. **Đồng bộ Database Pool**:
   * Loại bỏ việc tạo connection pool raw SQL độc lập. Sử dụng `db.DB()` từ GORM pool để truyền vào Batch Repository nhằm tối ưu hóa số lượng kết nối và hỗ trợ Transaction đồng nhất.
3. **Thêm Cơ Chế Rate Limiting & Blacklist**:
   * Bổ sung Middleware giới hạn tần suất request (Rate Limiting) trên Redis cho API gửi OTP để tránh brute-force.
   * Xây dựng cơ chế Blacklist Token khi người dùng Đăng xuất (Logout).
