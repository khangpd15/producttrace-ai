# Tài liệu Bàn giao: Tích hợp Dịch vụ Gửi Email (SendGrid & RabbitMQ)

## Tổng quan
Microservice NestJS (`nest-ai-service`) đã được cập nhật để hỗ trợ việc gửi email thông qua SendGrid. Hiện tại, service này đang chủ động lắng nghe sự kiện đăng ký người dùng từ RabbitMQ và sẽ tự động gửi email chào mừng/xác nhận khi nhận được sự kiện đó.

## Những công việc đã hoàn thành
1. **Tích hợp SendGrid (`MailModule` & `MailService`)**
   - Đã cài đặt và cấu hình `@sendgrid/mail` cùng với `@nestjs/config`.
   - Service tự động đọc các biến `SENDGRID_API_KEY`, `FROM_EMAIL`, và `TEMPLATE_ID` từ file `.env`.
   - Cung cấp các hàm `sendMail` (gửi văn bản thường/HTML) và `sendTemplateMail` (sử dụng dynamic template của SendGrid).

2. **RabbitMQ Consumer (`UserRegisteredConsumer`)**
   - Đã thêm một routing key mới là `user.registered` vào file `rabbitmq.constants.ts`.
   - Đã tạo `user-registered.consumer.ts` để lắng nghe sự kiện `user.registered`.
   - Tự động phân tích dữ liệu (payload) và gọi `MailService.sendTemplateMail()` với các dữ liệu nhận được từ message.

## Yêu cầu đối với Service Đăng ký (ví dụ: Go Core)
Để kích hoạt được luồng gửi email, service chịu trách nhiệm xử lý phần đăng ký người dùng cần phải gửi (publish) một message lên RabbitMQ với thông tin như sau:

- **Exchange**: Event exchange mặc định (ví dụ: `amq.topic` như được định nghĩa trong `.env`).
- **Routing Key**: `user.registered`
- **Định dạng Payload (JSON)**:
  ```json
  {
    "email": "user@example.com",
    "name": "Họ và Tên Người Dùng"
  }
  ```

Ngay khi message này được đẩy lên RabbitMQ, `nest-ai-service` sẽ tự động bắt lấy và thực hiện việc gửi email qua hệ thống SendGrid.

## Các biến môi trường cần thiết
Hãy đảm bảo rằng các biến dưới đây có mặt trong file `.env` khi chạy `nest-ai-service`:
```env
SENDGRID_API_KEY=mã_api_sendgrid_của_bạn_ở_đây
FROM_EMAIL=email_người_gửi_đã_được_xác_thực_ở_đây
TEMPLATE_ID=id_template_sendgrid_của_bạn_ở_đây
```

**Lưu ý bảo mật (Dành cho Team):**
- File `.env` chứa API Key nhạy cảm nên đã được đưa vào `.gitignore` và không đẩy lên Github.
- Thay vào đó, trong source code có sẵn file `.env.example`.
- Khi các thành viên khác lấy code về, vui lòng copy nội dung từ file `.env.example` tạo thành file `.env` trên máy local.
- Sau đó, hãy liên hệ trực tiếp với người cấu hình SendGrid (quản trị viên) qua tin nhắn nội bộ để lấy mã thật (`SENDGRID_API_KEY`, `FROM_EMAIL`, `TEMPLATE_ID`) và điền vào file `.env` của bạn. Mọi hành vi đẩy trực tiếp Key thật lên Git đều có rủi ro bị khóa tài khoản hoặc lợi dụng gửi email rác.
