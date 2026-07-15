# BÁO CÁO AUDIT DỰ ÁN: PRODUCTTRACE-AI
**Vai trò**: Senior Business Analyst + Solution Architect  
**Ngày thực hiện**: 15/07/2026  
**Trạng thái mã nguồn**: Chỉ Audit, không sửa đổi mã nguồn.

---

## 1. Go Core Service Audit (Module Inventory)

Go Core Service được xây dựng bằng Go (Sử dụng Gin Gonic và GORM kết nối PostgreSQL/Redis). Dưới đây là danh sách tất cả các module nghiệp vụ nằm trong thư mục `internal/modules` cùng trạng thái chi tiết của từng thành phần:

| Module | API (Handler) | Service | Repository | Database Entity | Trạng thái | Ghi chú / Chi tiết |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **audit** | ✅ Có (`audit_handler.go`) | ✅ Có (Ủy quyền qua `pkg/audit_log`) | ✅ Có (Ủy quyền qua `pkg/audit_log`) | ✅ Có (`AuditLog` entity / `audit_logs` table) | **Done** | Phục vụ ghi nhật ký hệ thống nâng cao. |
| **authen** | ✅ Có (`authen_handler.go`) | ✅ Có (`authen_service.go`) | ❌ Không cần (Dùng qua `userRepo` và Redis) | ❌ Không có riêng (Sử dụng bảng `users`) | **Done** | Cung cấp luồng đăng ký, đăng nhập, OTP và làm mới token qua Redis. |
| **batch** | ✅ Có (`batch_handler.go`) | ✅ Có (`batch_service.go`) | ✅ Có (`batch_repository.go`) | ✅ Có (`Batch` entity / `batches` table) | **Done** | Hỗ trợ tạo lô hàng, xuất danh sách sản phẩm, tạo file QR PDF tĩnh. |
| **dashboard** | ✅ Có (`dashboard_handler.go`) | ✅ Có (`dashboard_service.go`) | ✅ Có (`dashboard_repository.go`) | ❌ Không có (Truy vấn tổng hợp từ các bảng khác) | **Done** | Thống kê số lượng, hoạt động hệ thống, cảnh báo và biểu đồ bán hàng. |
| **event** | ❌ Thiếu (Không có router) | ⚠ Có (`event_service.go`) | ⚠ Có (`event_repository.go`) | ⚠ Có (`Event` entity / `events` table) | **In Progress** | File source đã viết xong nhưng chưa được khai báo ở `app.go` và `router.go`. Chưa hoạt động. |
| **location** | ✅ Có (`location_handler.go`) | ✅ Có (`location_service_iml.go`) | ✅ Có (`gorm_location_repository.go`) | ✅ Có (`Location` entity / `locations` table) | **Done** | Lưu địa điểm (Kho, đại lý, TT bảo hành). Cột `geo_location` sử dụng PostGIS. |
| **ownership** | ✅ Có (`ownership_handler.go`) | ✅ Có (`ownership_service.go`) | ✅ Có (`ownership_repo.go`) | ✅ Có (`Ownership` entity / `ownerships` table) | **Done** | Xử lý sở hữu chính thức, chuyển nhượng sở hữu sản phẩm giữa các khách hàng. |
| **product** | ✅ Có (`product_handler.go`) | ✅ Có (`product_service.go`) | ✅ Có (`product_repository.go`) | ✅ Có (`Product` entity / `products` table) | **Done** | Quản lý sản phẩm gốc, hỗ trợ cascade tạo/xóa các thành phần thuộc tính. |
| **product_category**| ✅ Có (`product_category_handler.go`)| ✅ Có (`product_category_service.go`)| ✅ Có (`product_category_repository.go`)| ✅ Có (`ProductCategory` / `product_categories`) | **Done** | Quản lý cây danh mục sản phẩm (hỗ trợ parent-child). |
| **product_variant** | ✅ Có (`product_variant_handler.go`)| ✅ Có (`product_variant_service.go`)| ✅ Có (`product_variant_repository.go`)| ✅ Có (`ProductVariant` / `product_variants`) | **Done** | Quản lý các biến thể sản phẩm vật lý (SKU, giá bán). |
| **product_attribute**| ✅ Có (`attribute_handler.go`) | ✅ Có (`attribute_service.go`) | ✅ Có (`attribute_repository.go`) | ✅ Có (`Attribute` entity / `attributes` table) | **Done** | Định nghĩa thuộc tính sản phẩm theo Category (như RAM, dung lượng). |
| **product_attribute_value**| ✅ Có (`attribute_value_handler.go`)| ✅ Có (`attribute_value_service.go`)| ✅ Có (`attribute_value_repository.go`)| ✅ Có (`AttributeValue` / `attribute_values`) | **Done** | Gán giá trị cụ thể cho từng thuộc tính của variant. |
| **product_item** | ✅ Có (`product_item_handler.go`)| ✅ Có (`product_item_service.go`)| ✅ Có (`product_item_repository.go`)| ✅ Có (`ProductItem` / `product_items`) | **Done** | Đại diện cho từng sản phẩm vật lý cụ thể (chứa mã code quét QR: `PTA-YYMM-XXXXXXXX`). |
| **public** | ✅ Có (`public_handler.go`) | ✅ Có (`public_service.go`) | ❌ Không có (Truy vấn qua các repo khác) | ❌ Không có | **Done** | API `/public/verify` phục vụ cho việc kiểm tra thông tin quét QR ẩn danh. |
| **trace** | ✅ Có (`trace_handler.go`) | ✅ Có (`trace_service.go`) | ✅ Có (`trace_repository.go`) | ✅ Có (`Trace` entity ánh xạ tới bảng `events`) | **Done** | Tra cứu lịch sử hành trình sản phẩm, xuất dữ liệu ra Excel/PDF. |
| **user** | ✅ Có (`user_handler.go`) | ✅ Có (`user_service.go`) | ✅ Có (`user_repository.go`) | ✅ Có (`User` entity / `users` table) | **Done** | Quản lý người dùng, phân quyền (ADMIN, STAFF, DEALER, CUSTOMER), khóa/mở tài khoản. |
| **warranty_claim** | ✅ Có (`warranty_claim_handler.go`)| ✅ Có (`warranty_claim_service.go`)| ✅ Có (`warranty_claim_repo.go`) | ✅ Có (`WarrantyClaim` / `warranty_claims` table) | **Done** | Người dùng gửi yêu cầu bảo hành. Tuy nhiên, Service đang dùng Mock cho Event/Audit/Notification ports. |

---

## 2. Nest AI Service Audit

Nest AI Service đảm nhận phần tương tác với Trí tuệ Nhân tạo, đồng bộ dữ liệu vectơ và xử lý các luồng bất đồng bộ (Event-Driven) thông qua RabbitMQ. Hệ thống cũng tích hợp một Microservice Python để sinh Embeddings.

### Liệt kê các module AI & Hạ tầng:

1. **Embedding Module (`src/modules/embedding`)**:
   - *Mô tả*: Nhận diện event sản phẩm từ RabbitMQ, tổng hợp văn bản thô, gọi Microservice Python để sinh vectơ nhúng 1024 chiều (BGE-M3), sau đó đẩy event `embedding.generated` trở lại hàng đợi.
   - *Đánh giá*: **Hoàn thiện (Done)**. Hoạt động bất đồng bộ tốt qua queue.
2. **Sync Module (`src/modules/sync`)**:
   - *Mô tả*: Consumer chuyên biệt để lắng nghe `embedding.generated`, tiến hành lưu trữ (upsert) vector cùng payload mô tả vào CSDL vector Qdrant. Khi hoàn tất sẽ phát event `embedding.completed`.
   - *Đánh giá*: **Hoàn thiện (Done)**.
3. **Search Module (`src/modules/search`)**:
   - *Mô tả*: Cung cấp API `POST /search/hybrid` để tìm kiếm lai (Hybrid Search). Hệ thống tạo vector truy vấn, áp dụng bộ lọc siêu dữ liệu (Metadata filter: category, manufacturer, province) gửi lên Qdrant, sau đó xếp hạng lại kết quả bằng thuật toán RRF (`RankingService`).
   - *Đánh giá*: **Hoàn thiện (Done)**.
4. **Reindex Module (`src/modules/reindex`)**:
   - *Mô tả*: Expose API `POST /reindex/all` để quét qua toàn bộ cơ sở dữ liệu sản phẩm trong Go Core (gọi qua HTTP client `ProductClientService`), sinh lại vector nhúng và nạp vào Qdrant.
   - *Đánh giá*: **Hoàn thiện (Done)**. Đã sửa lỗi trùng lặp `event_id` bằng timestamp suffix.
5. **Geo-Search Module (`src/modules/geo-search`)**:
   - *Mô tả*: Dự kiến xử lý tìm kiếm sản phẩm/địa điểm theo vị trí địa lý kết hợp vector (Geospatial AI Search).
   - *Đánh giá*: **Chưa có (Missing)**. Thư mục chỉ chứa file `readme.md` trống. Chưa viết dòng code nào.
6. **Recommendation Module (`src/modules/recommendation`)**:
   - *Mô tả*: Dự kiến đề xuất sản phẩm dựa trên độ tương đồng vector hoặc lịch sử mua sắm.
   - *Đánh giá*: **Chưa có (Missing)**. Thư mục chỉ chứa file `readme.md` trống.
7. **Qdrant Integration (`src/integrations/qdrant`)**:
   - *Mô tả*: Client giao tiếp trực tiếp với Qdrant. Tự động kiểm tra collection `product_embeddings`, cấu hình metric khoảng cách `Cosine` và kích thước vector cố định `1024`. Có cơ chế Singleton Promise chặn tình trạng tranh chấp (race condition) khi nhiều luồng cùng cố gắng tạo collection cùng lúc.
   - *Đánh giá*: **Hoàn thiện (Done)**.
8. **Queue / Messaging Service (`src/messaging`)**:
   - *Mô tả*: Trình quản lý RabbitMQ. Đăng ký các consumers xử lý email gửi đi như `notification.sent` (Bảo hành) hay `ownership.transferred` (Sở hữu) qua dịch vụ SendGrid.
   - *Đánh giá*: **Hoàn thiện (Done)**.
9. **Python Embedding Service (`apps/nest-ai-service/embedding-service`)**:
   - *Mô tả*: Viết bằng FastAPI và SentenceTransformers, nạp model `BAAI/bge-m3` trên CPU để sinh vector nhúng nhanh chóng.
   - *Đánh giá*: **Hoàn thiện (Done)**.

---

## 3. Đối chiếu với Business Flow (Luồng Nghiệp vụ)

| Business Flow | Backend Status | Frontend Status | Đánh giá Chung |
| :--- | :--- | :--- | :--- |
| **Product** (Quản lý & Đồng bộ sản phẩm) | ✅ **Hoàn thành**: Đầy đủ CRUD tại Go Core Service + Đồng bộ vector tự động tại Nest AI. | ❌ **Chưa có**: Hoàn toàn không có màn hình quản lý hay danh mục trên UI. | ⚠ **Thiếu đồng bộ**: Chỉ có BE, FE chưa phát triển. |
| **Timeline Scan** (Truy xuất nguồn gốc) | ✅ **Hoàn thành**: API public verify QR, xuất lịch sử sự kiện sản phẩm vật lý ra Excel/PDF. | ❌ **Chưa có**: Chưa có giao diện quét mã hay xem timeline hành trình sản phẩm. | ⚠ **Thiếu đồng bộ**: Chỉ có BE, FE chưa phát triển. |
| **Ownership** (Sở hữu & Chuyển nhượng) | ✅ **Hoàn thành**: Quy trình đăng ký OTP, xác thực sở hữu, chuyển nhượng sở hữu và gửi thông báo qua email. | ❌ **Chưa có**: Chưa phát triển Customer Portal để khách hàng quản lý và thực hiện luồng này. | ⚠ **Thiếu đồng bộ**: Chỉ có BE, FE chưa phát triển. |
| **Warranty** (Yêu cầu & Lịch sử bảo hành) | ⚠ **Thiếu**: Có API tạo yêu cầu bảo hành nhưng Service dùng mock adapter. Chưa có API GET lịch sử bảo hành cho khách hàng. | ❌ **Chưa có**: Chưa có màn hình gửi yêu cầu bảo hành hoặc tra cứu hạn bảo hành. | ❌ **Chưa hoàn thiện**: Backend thiếu API nghiệp vụ GET, Frontend chưa phát triển. |
| **AI Search** (Tìm kiếm thông minh lai) | ✅ **Hoàn thành**: Hoàn thiện API hybrid search trong Nest AI, hỗ trợ bộ lọc và chấm điểm xếp hạng. | ❌ **Chưa có**: Chưa có giao diện tìm kiếm thông minh cho người dùng cuối. | ⚠ **Thiếu đồng bộ**: Chỉ có BE, FE chưa phát triển. |

---

## 4. Kiểm tra Frontend Status

Hiện tại, Frontend (`web-frontend`) chỉ là một khung sườn Vite + React + Tailwind + React Router vô cùng đơn giản. Mức độ kết nối BE cực kỳ thấp.

### Chi tiết các trang:
- **Các trang đã kết nối với Backend**:
  - `ForgotPasswordPage.tsx`: Đã kết nối với Nest AI Service qua API `POST /auth/forgot-password` để gửi email chứa link đặt lại mật khẩu.
  - `ResetPasswordPage.tsx`: Đã kết nối với Nest AI Service qua API `GET /auth/validate-reset-token` (kiểm tra token từ email) và `POST /auth/reset-password` để cập nhật mật khẩu mới vào cơ sở dữ liệu giả lập (`users.json`).
- **Các trang còn mock/hardcode**:
  - `LoginPage.tsx`: Giao diện có sẵn nhưng hàm xử lý đăng nhập hoàn toàn bị mock cứng (`console.log('Login attempt', ...)`). Trang này chưa được kết nối với API đăng nhập thực tế của Go Core Service (`/api/auth/login`) hay Nest AI Service.
- **Các API chưa được Frontend sử dụng**:
  - **Tất cả** API của Go Core Service (Bao gồm: Auth/Login thực tế, CRUD Users, Lô hàng Batches, Danh mục và Sản phẩm Products, Locations, Dashboard hoạt động, Luồng Ownership đăng ký sở hữu, Yêu cầu bảo hành và luồng Timeline/Trace sản phẩm).
  - API Hybrid Search (`POST /search/hybrid`) và Reindex (`POST /reindex/all`) của Nest AI.
- **Component chưa hoàn thiện**:
  - Toàn bộ các component và màn hình cho luồng quản lý (Admin Dashboard), đại lý/kho hàng (Warehouse/Dealer View), và người tiêu dùng (Customer Portal) đều **chưa được xây dựng**.

---

## 5. Enterprise Gap (Các khoảng cách để đạt chuẩn Doanh nghiệp)

Để dự án ProductTrace-AI có thể vận hành thực tế ở mức độ doanh nghiệp, Solution Architect chỉ ra các khoảng trống công nghệ và nghiệp vụ lớn sau:

1. **Khoảng cách Frontend cực lớn**:
   - Giao diện người dùng thực chất mới chỉ có 2 trang Quên/Đặt lại mật khẩu chạy thực tế, còn lại toàn bộ nghiệp vụ lõi (Sản phẩm, QR Scan, Timeline, Bảo hành, Tìm kiếm AI, Quản trị) đều chưa được thiết kế trên frontend.
2. **Khoảng cách tích hợp Xác thực (Auth Integration Gap)**:
   - Go Core Service lưu trữ người dùng thật trong PostgreSQL với JWT token mã hóa trong khi Nest AI Service dùng cơ chế xác thực dựa trên database giả lập JSON (`users.json`). Chưa có cơ chế Single Sign-On (SSO) hoặc hệ thống phân phát JWT thống nhất giữa hai dịch vụ này.
3. **Thiếu API tra cứu thông tin bảo hành (Warranty GET API Gap)**:
   - Go Core Service chỉ có API tạo yêu cầu bảo hành (`POST /api/warranty-claims`). Hoàn toàn thiếu các API để khách hàng xem danh sách bảo hành của mình, xem hạn dùng hay cập nhật trạng thái bảo hành. Ngoài ra, module `warranty_claim` trong Go vẫn đang dùng mock ports cho các tính năng Event, Audit Log và Notification (chưa bắn message sang RabbitMQ thật).
4. **Khoảng trống tìm kiếm vị trí địa lý AI (Geospatial AI Search Gap)**:
   - Nest AI có PostGIS trong DB và Qdrant hỗ trợ lọc geo-spatial nhưng luồng tìm kiếm lai kết hợp bán kính địa điểm (`geo-search`) chưa được hiện thực hóa ở Nest AI (thư mục rỗng).
5. **Module Event trong Go Core chưa được đăng ký**:
   - Module `event` trong Go Core Service (chứa đầy đủ logic ghi nhận vết của sản phẩm) hiện không được khai báo trong router hay bootstrap của app. Điều này khiến luồng cập nhật lịch sử vị trí di chuyển thực tế của sản phẩm chưa được tự động kích hoạt đồng bộ.
6. **Môi trường và Cấu hình**:
   - Việc Nest AI đang sử dụng các file JSON tĩnh để lưu trữ token và người dùng quên mật khẩu gây mất an toàn thông tin và không thể scale ngang (load balancing) trong hạ tầng microservices của doanh nghiệp. Cần chuyển toàn bộ mock JSON sang PostgreSQL/Prisma.

---

## 6. Kết luận Mức độ Hoàn thiện

Dựa trên phân tích sâu về mã nguồn và luồng nghiệp vụ của toàn bộ dự án:

*   **Mức độ hoàn thiện Backend (Go Core + Nest AI)**: **B** (Hấu hết các API cốt lõi, RabbitMQ message brokers, Qdrant integration, Python embeddings đã viết rất tốt và hoạt động ổn định, chỉ thiếu một vài tích hợp nâng cao và wiring).
*   **Mức độ hoàn thiện Frontend**: **E** (Gần như chưa phát triển, chỉ có khung sườn trống và 2 trang quên mật khẩu).
*   **Mức độ hoàn thiện Tổng thể Dự án**: **D** (Hệ thống backend mạnh mẽ nhưng dự án chưa thể sử dụng hay đưa vào demo/vận hành vì thiếu hoàn toàn giao diện người dùng cho các luồng nghiệp vụ cốt lõi).

---
*Tài liệu được sinh tự động bởi Senior Business Analyst + Solution Architect. Không có dòng code nào bị thay đổi.*
