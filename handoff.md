# Tài liệu Bàn giao: Module Quên/Đổi Mật khẩu (Fullstack) & SendGrid

## Tổng quan
Dự án đã được tích hợp hoàn chỉnh luồng **Quên Mật Khẩu (Forgot Password)** từ Backend (NestJS) tới Frontend (React/Vite). Toàn bộ hệ thống gửi email qua SendGrid và xác thực mật khẩu đều đã hoạt động trơn tru. Backend hiện đang sử dụng cơ sở dữ liệu giả lập (Mock JSON) với thiết kế chuẩn Clean Architecture để dễ dàng thay đổi sang Prisma/PostgreSQL sau này.

---

## 1. Backend (NestJS - `nest-ai-service`)

### Kiến trúc & Database (Mock JSON)
- Không sử dụng CSDL thật để dễ dàng test. Dữ liệu được lưu trực tiếp vào 2 file JSON đóng vai trò như database:
  - `src/mock-data/users.json` (Lưu thông tin User và Mật khẩu băm - `bcrypt`)
  - `src/mock-data/password-reset-tokens.json` (Lưu Token quên mật khẩu)
- Áp dụng **Repository Pattern**: `AuthService` hoàn toàn giao tiếp qua các Interface (`UserRepository`, `PasswordResetRepository`). Khi tích hợp database thật, Dev chỉ cần viết class `PrismaUserRepository` và thay đổi DI provider trong `AuthModule`. Không cần sửa đổi logic nghiệp vụ.

### Các API đã hoàn thiện
1. **`POST /auth/forgot-password`**
   - **Body**: `{ "email": "test@gmail.com" }`
   - Tạo token bảo mật bằng `crypto.randomBytes()`.
   - Lưu trữ mã băm (hash) của token (chuẩn bảo mật), có hạn 15 phút.
   - Gọi `MailService` để gửi SendGrid email (có chứa link tới Frontend).
   - *Lưu ý*: API luôn trả về `200 OK` kể cả email không tồn tại để chống lại lỗ hổng dò quét email (Email Enumeration).

2. **`GET /auth/validate-reset-token`**
   - **Query**: `?token=...&email=...`
   - Kiểm tra tính hợp lệ và thời hạn của token. API này được Frontend tự động gọi khi User mở link trong email.

3. **`POST /auth/reset-password`**
   - **Query**: `?email=...` | **Body**: `{ "token": "...", "password": "...", "confirmPassword": "..." }`
   - Kiểm tra token một lần nữa.
   - Mã hoá (hash) mật khẩu mới bằng `bcrypt`.
   - Update mật khẩu trong JSON và xoá token đó đi để đảm bảo Link chỉ dùng được 1 lần duy nhất.

---

## 2. Frontend (React + Vite + TailwindCSS - `web-frontend`)

### Giao diện & Luồng xử lý
- Đã khôi phục hoàn chỉnh cấu trúc Vite/React bị thiếu trước đó (`package.json`, `vite.config.ts`, `main.tsx`, `App.tsx`, cài đặt TailwindCSS...).
- Đã thiết kế 3 trang chính dùng React Router:
  1. `/login`: Giao diện đăng nhập.
  2. `/forgot-password`: Màn hình điền email để yêu cầu đổi mật khẩu.
  3. `/reset-password`: Màn hình thay đổi mật khẩu. Đã xử lý kỹ lưỡng UX/UI:
     - Tự động gọi API kiểm tra Token (`cache: 'no-store'`) để tránh lỗi cache trình duyệt khi click link nhiều lần.
     - Kiểm tra sức mạnh mật khẩu.
     - Hiển thị màn hình Lỗi/Thành công rõ ràng, tránh hiển thị form nếu Token không hợp lệ.

### Cách chạy Frontend
```bash
cd apps/web-frontend
npm install
npm run dev
```
Trang web sẽ chạy ở `http://localhost:5173`.

---

## 3. Cấu hình Môi trường (.env)

Đảm bảo file `.env` ở thư mục root có đầy đủ các biến sau:

```env
# URL của Frontend để NestJS nhúng vào nội dung Email SendGrid
FRONTEND_URL=http://localhost:5173

# Cấu hình SendGrid
SENDGRID_API_KEY=mã_api_sendgrid_của_bạn_ở_đây
FROM_EMAIL=email_người_gửi_đã_được_xác_thực_ở_đây
WELCOME_TEMPLATE_ID=id_template_chào_mừng_ở_đây
RESET_PASSWORD_TEMPLATE_ID=id_template_quên_mật_khẩu_ở_đây
```

**Chú ý về Template HTML của SendGrid:**
- File template HTML dành cho SendGrid nằm tại `reset-password-email.html` (đã được Dev cung cấp với thiết kế bảng tương thích tốt trên mọi nền tảng).
- Template có các biến: `{{name}}`, `{{resetLink}}`, `{{year}}`, `{{companyName}}`. 
- NestJS đã được cấu hình truyền đủ dữ liệu cho các biến này. Riêng `{{companyName}}` bạn có thể hardcode thẳng vào cấu hình SendGrid.

## 4. Hướng dẫn Test (Cho Tester/Dev khác)
1. Đảm bảo chạy cả NestJS (`npm run start:dev`) và Vite Frontend (`npm run dev`).
2. Mở file `apps/nest-ai-service/src/mock-data/users.json` và điền email thật của bạn vào (để nhận được thư).
3. Mở trình duyệt vào `http://localhost:5173/login`, bấm "Forgot your password?".
4. Nhập email và submit.
5. Check hộp thư đến, mở email và click vào nút "Xác nhận thay đổi mật khẩu".
6. Điền mật khẩu mới trên form hiện ra.
7. Reset thành công -> Kiểm tra file `users.json` sẽ thấy chuỗi băm `password` đã bị thay đổi, và `password-reset-tokens.json` đã rỗng. Mọi thứ hoạt động như một hệ thống thực thụ!

---

# Tài liệu Bàn giao: Module Thông Báo Bảo Hành (Warranty Notification)

## Tổng quan
Đã triển khai hoàn chỉnh tính năng gửi **email thông báo cập nhật trạng thái bảo hành** cho khách hàng thông qua hệ thống Event-Driven hiện có (RabbitMQ → NestJS Worker → SendGrid). Tính năng nằm trên nhánh `feature/notification-warranty`.

- **Use Case**: `UC-P3-NOTI-01` — Gửi thông báo cập nhật bảo hành tới email khách hàng khi trạng thái thay đổi.
- **Nhánh Git**: `feature/notification-warranty`
- **SendGrid Template ID**: `d-aa9b56ba4bf64b54a72eddc7ba33ba03`

---

## 1. Các file đã thay đổi

| File | Mô tả thay đổi |
|---|---|
| `src/messaging/rabbitmq/rabbitmq.constants.ts` | Thêm `NOTIFICATION_SENT: 'notification.sent'` vào `ROUTING_KEYS` và `EVENT_TYPES` |
| `src/messaging/rabbitmq/rabbitmq.service.ts` | Thêm `NOTIFICATION_SENT` vào danh sách `routingKeys` để tự động bind hàng đợi khi khởi tạo |
| `src/messaging/consumers/notification.consumer.ts` | Thêm 3 trường động vào `NotificationPayload` và xử lý `case notification.sent` trong switch |
| `src/mail/mail.service.ts` | Thêm/Cập nhật method `sendWarrantyUpdateEmail(to, fullName, productName, status, endDate)` hỗ trợ đầy đủ các tham số động vào SendGrid Template hoặc fallback HTML |
| `.env.example` | Thêm biến `WARRANTY_UPDATE_TEMPLATE_ID` |

---

## 2. Luồng hoạt động (Event Flow)

```
Go Core Service (bảo hành update)
    │
    │  Publish event: routing_key="notification.sent"
    │  Payload: { email, full_name, product_name, warranty_status, warranty_end_date }
    ▼
RabbitMQ Queue: "ai.events"
    │
    ▼
NotificationConsumer (NestJS Worker)  ← lắng nghe liên tục
    │  case "notification.sent"
    │  Trích xuất payload động
    ▼
MailService.sendWarrantyUpdateEmail()
    │  Gọi SendGrid API với templateId + dynamicTemplateData
    ▼
Email đến hòm thư khách hàng  ✅
```

---

## 3. Cấu trúc Payload RabbitMQ

Khi service bảo hành (Go Core Service) publish event, JSON message phải có cấu trúc:

```json
{
  "event_type": "notification.sent",
  "data": {
    "email": "khachhang@gmail.com",
    "full_name": "Nguyễn Văn A",
    "product_name": "iPhone 15 Pro Max 256GB",
    "warranty_status": "Đã hoàn tất sửa chữa",
    "warranty_end_date": "24/10/2026"
  }
}
```

> **Lưu ý**: Tất cả 5 trường trên đều là dữ liệu thật (dynamic), không hardcode. Nếu thiếu trường nào, hệ thống có giá trị mặc định fallback để không crash.

---

## 4. Cấu hình SendGrid Template

**Template ID**: `d-aa9b56ba4bf64b54a72eddc7ba33ba03`

Các biến (variables) được NestJS truyền vào template:

| Tên biến SendGrid | Dữ liệu truyền vào | Ví dụ |
|---|---|---|
| `{{fullName}}` | Tên khách hàng | `Nguyễn Văn A` |
| `{{productName}}` | Tên sản phẩm | `iPhone 15 Pro Max 256GB` |
| `{{status}}` | Trạng thái bảo hành | `Đã hoàn tất sửa chữa` |
| `{{endDate}}` | Ngày hết hạn bảo hành | `24/10/2026` |
| `{{frontendUrl}}` | Link hệ thống | `http://localhost:5173` |
| `{{year}}` | Năm hiện tại | `2025` |

---

## 5. Cấu hình môi trường (.env)

Thêm biến sau vào file `.env` của `apps/nest-ai-service`:

```env
# SendGrid Template ID cho thông báo bảo hành
WARRANTY_UPDATE_TEMPLATE_ID=d-aa9b56ba4bf64b54a72eddc7ba33ba03
```

> Nếu biến này **không được set** hoặc để trống, hệ thống sẽ tự động dùng Template ID `d-aa9b56ba4bf64b54a72eddc7ba33ba03` làm giá trị mặc định (đã hardcode trong code).
> Nếu cả SendGrid API Key cũng không được cấu hình, hệ thống sẽ chạy ở **MOCK mode** và chỉ log ra console mà không gửi email thật.

---

## 6. Hướng dẫn Test

### Cách A: Bắn event thủ công qua RabbitMQ Management UI
1. Mở `http://localhost:15672` (đăng nhập: `admin` / `admin123`).
2. Vào **Queues** → chọn queue `ai.events`.
3. Tìm mục **Publish message**, điền:
   - **Routing key**: `notification.sent`
   - **Payload**: JSON như mục 3 ở trên
4. Nhấn **Publish message**.
5. Kiểm tra log của `nest-ai-service` sẽ thấy:
   ```
   [NotificationConsumer] Warranty update email sent to khachhang@gmail.com
   ```
6. Kiểm tra hòm thư nhận email với giao diện từ template SendGrid.

### Cách B: Chạy script test trực tiếp
File test được tạo sẵn tại `apps/nest-ai-service/test-email.js`. Script này đã được tối ưu để tự động đọc cấu hình SendGrid (`SENDGRID_API_KEY`, `SENDGRID_FROM_EMAIL`, `WARRANTY_UPDATE_TEMPLATE_ID`) từ file `.env`.

Cách chạy:
```bash
cd apps/nest-ai-service
node test-email.js
```

**Nhật ký kiểm thử (Mới nhất):**
- **Ngày thực hiện:** 14/07/2026
- **Email nhận test:** `nguyenzc17@gmail.com` (Đã được cấu hình làm email nhận mặc định trong script để chạy kiểm thử).
- **Kết quả:** Gửi thành công (`✅ GỬI EMAIL THÀNH CÔNG!`). API SendGrid phản hồi status 2xx, email được chuyển tiếp đến hàng đợi phân phát của SendGrid đến hộp thư đích.

---

## 7. Điểm quan trọng cho Dev tiếp nhận

- **Không cần sửa gì thêm** trong NestJS khi Go Core Service thay đổi nội dung bảo hành — chỉ cần đảm bảo payload JSON đúng schema như mục 3.
- Để thêm event thông báo loại mới (ví dụ: `warranty.expired`), chỉ cần thêm:
  1. Constant mới vào `rabbitmq.constants.ts`
  2. `case` mới vào `notification.consumer.ts`
  3. Method mới vào `mail.service.ts`
- **Template fallback**: Nếu chưa setup SendGrid Template, hệ thống tự động gửi email HTML thuần có đầy đủ thông tin (không bị lỗi).


