# Tài liệu Danh sách API (Endpoints) - ProductTrace AI

Tài liệu này thống kê chi tiết toàn bộ các API endpoints được định nghĩa trong hệ thống, chủ yếu từ file [router.go](file:///D:/producttrace-ai/apps/go-core-service/internal/router/router.go) và các route hệ thống khác của `go-core-service`.

---

## 1. Cấu hình Chung & Middleware Toàn Cục

Hệ thống sử dụng **Gin Web Framework** cho phần core service. Tất cả các endpoint (trừ các route test/health của hệ thống) đều nằm dưới nhóm tiền tố (prefix) `/api`.

### Middleware Toàn cục (Global Middlewares):
- **CORS Middleware**: Cho phép các request từ frontend chạy tại `http://localhost:3000` và `http://localhost:5173`. Hỗ trợ các method: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`.
- **Recovery Middleware**: Tự động phục hồi khi có panic xảy ra trong quá trình xử lý request, đảm bảo service không bị crash.
- **RequestID Middleware**: Tự động đính kèm một `X-Request-ID` độc nhất vào header của mỗi request để dễ dàng trace log.
- **Logger Middleware**: Ghi nhận thông tin của mọi request (thời gian, status code, IP, method, path).

### Phục vụ File Tĩnh (Static Files):
- **`GET /storage`**: Cấu hình phục vụ các file tĩnh trong thư mục `./storage` (ví dụ: các chứng chỉ, hình ảnh xuất ra).

---

## 2. Danh sách Endpoints chi tiết theo Module

### 2.1. Module Xác thực (Authentication)
* **Tiền tố nhóm:** `/api/auth`
* **Handler:** `AuthHandler` ([authen_handler.go](file:///D:/producttrace-ai/apps/go-core-service/internal/modules/authen/handler/authen_handler.go))
* **Bảo mật:** Tất cả các endpoint trong nhóm này đều là **Công khai (Public)**.

| HTTP Method | Path | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/auth/register` | `ah.Register` | Đăng ký tài khoản người dùng mới (yêu cầu xác thực OTP sau đó). |
| `POST` | `/api/auth/login` | `ah.Login` | Đăng nhập tài khoản bằng Email và Mật khẩu. Trả về Access Token và Refresh Token. |
| `POST` | `/api/auth/verify-otp` | `ah.VerifyOTP` | Xác thực tài khoản bằng mã OTP được gửi qua Email. |
| `POST` | `/api/auth/resend-otp` | `ah.ResendOTP` | Gửi lại mã OTP xác thực tài khoản vào Email. |
| `POST` | `/api/auth/refresh` | `ah.RefreshToken` | Làm mới Access Token sử dụng Refresh Token hợp lệ. |
| `POST` | `/api/auth/logout` | `ah.Logout` | Đăng xuất khỏi hệ thống và vô hiệu hóa Refresh Token hiện tại. |
| `POST` | `/api/auth/forgot-password` | `ah.ForgotPassword` | Yêu cầu khôi phục mật khẩu. Hệ thống gửi mã OTP khôi phục qua Email. |
| `POST` | `/api/auth/reset-password` | `ah.ResetPassword` | Đặt lại mật khẩu mới sử dụng mã OTP khôi phục đã nhận. |

---

### 2.2. Module Người dùng (User Management)
* **Tiền tố nhóm:** `/api/users`
* **Handler:** `UserHandler`
* **Bảo mật:** Yêu cầu đăng nhập (`AuthMiddleware`) và phân quyền vai trò cụ thể.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/users/profile` | Đăng nhập (`AuthMiddleware`) | `uh.GetProfile` | Xem thông tin hồ sơ của cá nhân đang đăng nhập. |
| `PUT` | `/api/users/profile/:id` | Đăng nhập (`AuthMiddleware`) | `uh.UpdateProfile` | Cập nhật hồ sơ cá nhân theo ID người dùng. |
| `POST` | `/api/users` | Admin (`ADMIN`) | `uh.CreateUser` | Tạo tài khoản người dùng mới trực tiếp (thường dùng cho quản trị viên). |
| `PUT` | `/api/users/:id` | Admin (`ADMIN`) | `uh.UpdateUser` | Cập nhật thông tin chi tiết của người dùng bất kỳ theo ID. |
| `DELETE` | `/api/users/:id` | Admin (`ADMIN`) | `uh.DeleteUser` | Xóa người dùng khỏi hệ thống theo ID. |
| `GET` | `/api/users` | Admin (`ADMIN`) | `uh.ListUsers` | Lấy danh sách toàn bộ người dùng trong hệ thống (hỗ trợ phân trang/tìm kiếm). |
| `GET` | `/api/users/:id` | Admin (`ADMIN`) | `uh.GetUserDetail` | Xem chi tiết thông tin của một người dùng cụ thể theo ID. |

---

### 2.3. Module Lô sản phẩm (Batch Management)
* **Tiền tố nhóm:** `/api/batches`
* **Handler:** `BatchHandler` ([batch_handler.go](file:///D:/producttrace-ai/apps/go-core-service/internal/modules/batch/handler/batch_handler.go))
* **Bảo mật:** Gồm cả endpoint công khai và bảo vệ nhiều lớp quyền hạn khác nhau.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/batches/:id` | **Công khai** | `bh.GetBatchDetail` | Xem chi tiết lô sản phẩm bằng mã hoặc ID lô. Endpoint này mở công khai để khách hàng quét mã QR truy xuất nguồn gốc. |
| `GET` | `/api/batches` | Đăng nhập (`AuthMiddleware`) | `bh.GetBatchList` | Lấy danh sách lô hàng. Đối với người dùng thông thường không phải Admin/Staff, hệ thống ẩn các lô có trạng thái nháp (`DRAFT`). |
| `GET` | `/api/batches/search` | Đăng nhập (`AuthMiddleware`) | `bh.SearchBatch` | Tìm kiếm lô sản phẩm dựa trên các tiêu chí lọc (mã, tên, trạng thái...). |
| `GET` | `/api/batches/:id/events` | Đăng nhập (`AuthMiddleware`) | `bh.GetBatchEvents` | Lấy danh sách tất cả các sự kiện thay đổi, di chuyển hoặc kiểm tra liên quan đến lô hàng đó. |
| `GET` | `/api/batches/:id/products` | Đăng nhập & `ADMIN`, `STAFF`, `DEALER` | `bh.GetBatchProducts` | Xem danh sách các sản phẩm (vật phẩm cụ thể) thuộc lô hàng. |
| `GET` | `/api/batches/:id/history` | Đăng nhập & `ADMIN`, `STAFF` | `bh.GetBatchHistory` | Xem lịch sử thay đổi/audit log chi tiết của lô hàng. Không cho phép vai trò `DEALER` hay `CUSTOMER` truy cập. |
| `POST` | `/api/batches/:id/export` | Đăng nhập & `ADMIN`, `MANAGER`, `WAREHOUSE` | `bh.ExportBatch` | Thực hiện xuất kho/chuyển lô hàng đến vị trí khác. |
| `GET` | `/api/batches/export-qr/:id` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `bh.ExportQR` | Xuất file PDF chứa danh sách các mã QR của lô hàng để in ấn và dán nhãn vật lý. |
| `POST` | `/api/batches` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `bh.CreateBatch` | Tạo mới một lô sản phẩm (khởi tạo trạng thái ban đầu). |
| `PATCH` | `/api/batches/:id/status` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `bh.UpdateBatchStatus` | Cập nhật nhanh trạng thái của lô sản phẩm (ví dụ: chuyển từ Nháp sang Đã kích hoạt). |
| `DELETE` | `/api/batches/:id` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `bh.DeleteBatch` | Xóa lô sản phẩm theo ID. |

---

### 2.4. Module Sản phẩm (Product Management)
* **Tiền tố nhóm:** `/api/products`
* **Handler:** `ProductHandler`
* **Bảo mật:** Cho phép xem công khai nhưng giới hạn quyền chỉnh sửa/thêm mới.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/products` | **Công khai** | `ph.GetAllProducts` | Lấy danh sách toàn bộ danh mục sản phẩm có trong hệ thống. |
| `GET` | `/api/products/:id` | **Công khai** | `ph.GetProductByID` | Xem thông tin chi tiết của một loại sản phẩm cụ thể. |
| `POST` | `/api/products` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `ph.CreateProduct` | Tạo mới một loại sản phẩm vào danh mục. |
| `PUT` | `/api/products/:id` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `ph.UpdateProduct` | Cập nhật thông tin chi tiết loại sản phẩm hiện có. |
| `DELETE` | `/api/products/:id` | Đăng nhập & `ADMIN` | `ph.DeleteProduct` | Xóa một loại sản phẩm khỏi danh mục hệ thống. |

---

### 2.5. Module Vị trí địa lý (Location Management)
* **Tiền tố nhóm:** `/api/locations`
* **Handler:** `LocationHandler`
* **Bảo mật:** Cho phép xem danh sách công khai, chỉ có quản trị viên mới được cấu hình.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/locations` | **Công khai** | `locationHandler.GetAll` | Lấy danh sách toàn bộ các địa điểm/kho bãi lưu trữ hoặc nhà máy. |
| `GET` | `/api/locations/:id` | **Công khai** | `locationHandler.GetByID` | Xem thông tin chi tiết của một địa điểm cụ thể. |
| `POST` | `/api/locations` | Đăng nhập & `ADMIN` | `locationHandler.Create` | Tạo mới một vị trí/địa điểm hoạt động. |
| `PUT` | `/api/locations/:id` | Đăng nhập & `ADMIN` | `locationHandler.Update` | Cập nhật thông tin chi tiết của vị trí đã tồn tại. |
| `DELETE` | `/api/locations/:id` | Đăng nhập & `ADMIN` | `locationHandler.Delete` | Xóa thông tin vị trí khỏi hệ thống. |

---

### 2.6. Module Bảng điều khiển (Dashboard Stats)
* **Tiền tố nhóm:** `/api/dashboard`
* **Handler:** `DashboardHandler`
* **Bảo mật:** Yêu cầu quyền vận hành quản trị hoặc nhân sự hệ thống.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/dashboard/stats` | Đăng nhập & `ADMIN`, `STAFF` | `dh.GetStats` | Thống kê số liệu hoạt động của hệ thống (lô hàng, sản phẩm, thiết bị quét, v.v.). |

---

### 2.7. Module Phiên bản sản phẩm (Product Variant)
* **Tiền tố nhóm:** `/api/variants`
* **Handler:** `ProductVariantHandler`
* **Bảo mật:** Các thông tin xem là công khai, chỉnh sửa cần quyền Nhà sản xuất hoặc Admin.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/variants/:id` | **Công khai** | `vh.GetVariantByID` | Lấy chi tiết thông tin của một phiên bản (variant) cụ thể. |
| `GET` | `/api/variants/product/:product_id` | **Công khai** | `vh.GetVariantsByProductID` | Lấy toàn bộ các biến thể thuộc về một ID sản phẩm chính. |
| `PUT` | `/api/variants/:id` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `vh.UpdateVariant` | Cập nhật thông tin chi tiết một phiên bản sản phẩm. |
| `DELETE` | `/api/variants/:id` | Đăng nhập & `ADMIN` | `vh.DeleteVariant` | Xóa phiên bản sản phẩm khỏi hệ thống. |

---

### 2.8. Module Thuộc tính sản phẩm (Product Attribute)
* **Tiền tố nhóm:** `/api/attributes`
* **Handler:** `AttributeHandler`
* **Bảo mật:** Cho phép duyệt xem công khai, thao tác tạo/sửa/xóa yêu cầu quyền cao hơn.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/attributes` | **Công khai** | `ah.ListAttributes` | Lấy danh sách các loại thuộc tính sản phẩm hiện có (ví dụ: Size, Màu sắc, Chất liệu...). |
| `GET` | `/api/attributes/:id` | **Công khai** | `ah.GetAttributeByID` | Lấy chi tiết thông tin của một thuộc tính cụ thể theo ID. |
| `POST` | `/api/attributes` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `ah.CreateAttribute` | Khai báo thuộc tính sản phẩm mới. |
| `PUT` | `/api/attributes/:id` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `ah.UpdateAttribute` | Cập nhật thông tin thuộc tính hiện có. |
| `DELETE` | `/api/attributes/:id` | Đăng nhập & `ADMIN` | `ah.DeleteAttribute` | Xóa thuộc tính khỏi hệ thống. |

---

### 2.9. Module Giá trị Thuộc tính (Product Attribute Value)
* **Tiền tố nhóm:** `/api/attribute-values` và `/api/variants` (được đăng ký chung trong `SetupProductAttributeValueRouter`)
* **Handler:** `AttributeValueHandler`

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/attribute-values` | **Công khai** | `ah.ListAllAttributeValues` | Liệt kê toàn bộ các giá trị thuộc tính đã lưu. |
| `GET` | `/api/attribute-values/:id` | **Công khai** | `ah.GetAttributeValueByID` | Lấy chi tiết thông tin một giá trị thuộc tính cụ thể theo ID. |
| `PUT` | `/api/attribute-values/:id` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `ah.UpdateAttributeValue` | Cập nhật giá trị thuộc tính. |
| `DELETE` | `/api/attribute-values/:id` | Đăng nhập & `ADMIN` | `ah.DeleteAttributeValue` | Xóa giá trị thuộc tính khỏi hệ thống. |
| `GET` | `/api/variants/:id/attributes` | **Công khai** | `ah.GetAttributeValuesByVariantID` | Xem danh sách các thuộc tính và giá trị tương ứng của một phiên bản cụ thể. |
| `POST` | `/api/variants/:id/attributes` | Đăng nhập & `ADMIN`, `MANUFACTURER` | `ah.AssignAttributes` | Gán các giá trị thuộc tính cho một phiên bản sản phẩm cụ thể. |

---

### 2.10. Module Truy vết hành trình (Trace / Timeline)
* **Tiền tố nhóm:** `/api/trace`
* **Handler:** `TraceHandler` ([trace_handler.go](file:///D:/producttrace-ai/apps/go-core-service/internal/modules/trace/handler/trace_handler.go))

| HTTP Method | Path | Yêu cầu Quyền hạn / Giới hạn | Handler Function | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/trace/search` | **Công khai** (Giới hạn: 30 requests/phút) | `th.Search` | Tìm kiếm hành trình sản phẩm (Timeline) dựa trên mã sản phẩm hoặc số Serial. Có áp dụng Rate Limiter để chống Spam. |
| `POST` | `/api/trace/export/pdf` | Đăng nhập (`AuthMiddleware`) | `th.ExportPDF` | Xuất hành trình sản phẩm ra định dạng PDF làm tài liệu đối chiếu hoặc chứng nhận. |
| `POST` | `/api/trace/export/excel` | Đăng nhập (`AuthMiddleware`) | `th.ExportExcel` | Xuất hành trình sản phẩm ra bảng tính Excel phục vụ phân tích báo cáo. |

---

### 2.11. Các Endpoints Tiện ích Hệ thống (System Services)
Các endpoint này được cấu hình trực tiếp ở cấp khởi tạo hệ thống trong [main.go](file:///D:/producttrace-ai/apps/go-core-service/cmd/api/main.go) phục vụ kiểm tra hệ thống và liên kết dịch vụ nền.

| HTTP Method | Path | Yêu cầu Quyền hạn | Handler / Mô tả |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | **Công khai** | Kiểm tra trạng thái hoạt động (Health check) của `go-core-service`. Trả về: `{"service": "go-core-service", "status": "ok"}`. |
| `GET` | `/test-event` | **Công khai** | Route test để phát sinh (Publish) một sự kiện thử nghiệm (`ProductCreated`) lên hệ thống tin nhắn RabbitMQ. |
