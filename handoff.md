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
