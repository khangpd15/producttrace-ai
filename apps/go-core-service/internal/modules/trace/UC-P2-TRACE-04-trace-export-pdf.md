# UC-P2-TRACE-04 - Xuất PDF lịch sử truy vết

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Traceability |
| Use Case ID | UC-P2-TRACE-04 |
| Feature Name | Xuất PDF lịch sử truy vết (PDF Export API) |
| Priority | Medium |
| Git Branch | feature/trace-export-pdf |
| Producer | trace.exported |
| Consumer | Notification |
| Dependency | UC-P1-TRACE-03 (View Event List) |

## Description

- **Business Purpose**: Hỗ trợ xuất hành trình vòng đời (Timeline) của sản phẩm ra một file tài liệu PDF sang trọng, có cấu trúc thẩm mỹ cao, phù hợp để in ấn kẹp vào hồ sơ lô hàng, dán kèm lên thùng carton lớn hoặc gửi qua Email cho đối tác làm chứng chỉ nguồn gốc xuất xứ.
- **User Problem Solved**: Khách hàng doanh nghiệp yêu cầu một bản in vật lý chứng thực nguồn gốc từ nhà sản xuất. Nhân viên không thể chụp màn hình giao diện web vì trông thiếu chuyên nghiệp. Họ cần một tệp tài liệu PDF chuẩn hóa có logo, bảng biểu rõ ràng và chữ ký điện tử xác nhận của hệ thống.
- **Expected System Behavior**: Hệ thống tiếp nhận ID của sản phẩm cần kết xuất, truy xuất toàn bộ thông tin sản phẩm và mảng timeline sự kiện, sử dụng thư viện render PDF (Ví dụ: `PDFKit` hoặc `puppeteer` chụp trang in) để dựng tài liệu PDF có đầy đủ header/footer, lưu tệp vào Cloud Storage, phát đi RabbitMQ event `trace.exported` để gửi thông báo (Notification) kèm link tải trực tiếp đến cho User.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Admin / Supervisor | Xuất file PDF để làm thủ tục chứng nhận CO/CQ (Certificate of Origin/Quality) cho hàng xuất khẩu. |
| Staff Kho | Xuất in nhãn đính kèm biên bản giao nhận hàng hóa cho tài xế xe tải. |
| Notification Service| Consumer nhận event để gửi Email hoặc In-app notification báo tệp PDF đã sẵn sàng tải về. |

---

# 3. Permission

| Role | Create | Update | Delete | View |
|---|---|---|---|---|
| Admin | Yes | No | No | Yes |
| Staff Kho | Yes | No | No | Yes |
| Dealer | Yes | No | No | Yes |
| Customer | No | No | No | Yes |

---

# 4. Preconditions

- Người dùng đã đăng nhập và được cấp quyền xuất tài liệu hệ thống (`EXPORT_TRACE_PDF`).
- Sản phẩm và timeline liên quan tồn tại trong hệ thống.

---

# 5. Trigger

User nhấp chọn nút "Export PDF" (Biểu tượng tài liệu PDF màu đỏ) tại giao diện xem Timeline sản phẩm.

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại màn hình chi tiết hành trình, nhấn nút "Export PDF" và xác nhận. | Hiển thị thông báo: "Generating PDF. This might take a few seconds... We will notify you when it's done." và gửi request POST sang API. |
| 2 | System | Tiếp nhận Yêu cầu | Đăng ký một Job xử lý ngầm (Asynchronous job) để tránh nghẽn luồng xử lý chính nếu tệp dữ liệu lớn. |
| 3 | System | Thu thập Dữ liệu | Truy vấn PostgreSQL lấy thông tin `product_items`, `products`, `batches` và toàn bộ mảng `events` liên kết. |
| 4 | System | Dựng layout PDF | Truyền dữ liệu vào Template in ấn chuẩn hóa: có logo ProductTrace AI, bảng thông tin kỹ thuật, vẽ hình sơ đồ chuỗi các mốc sự kiện dọc tinh tế. |
| 5 | System | Lưu trữ File | Generate luồng nhị phân (Binary stream) thành file `.pdf`, upload lên Cloud Storage (hoặc thư mục tệp tạm bảo mật) và sinh link rút gọn có thời hạn (Signed URL). |
| 6 | System | Gửi Event RabbitMQ | Publish tin nhắn `trace.exported` loại `PDF` lên RabbitMQ exchange `trace.exports`. |
| 7 | System | Ghi Audit Log | Ghi lại vết hành động xuất file của User. |
| 8 | System | Trả kết quả | Phản hồi lại mã 202 Accepted (Yêu cầu đã được tiếp nhận) cho Frontend. |

---

# 7. Alternative Flow

## AF-001 Tải về trực tiếp (Synchronous Download for small data)
- **Description**: Nếu timeline của sản phẩm ngắn (Dưới 10 sự kiện), hệ thống sẽ thực hiện generate PDF đồng bộ (Synchronous) và trả thẳng stream file trực tiếp về trình duyệt trong vòng 2 giây, giúp người dùng lưu tệp tức thì mà không cần đợi chạy job ngầm.

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Thất bại khi upload file PDF lên Cloud Storage | Trả về lỗi 500: "Export failed due to storage connection issues. Please try again later." |
| ERR-002 | Timeline sản phẩm rỗng (Chưa có bất kỳ mốc sự kiện nào) | Chặn xuất và trả lỗi: "Cannot export PDF for a product with an empty timeline." |

---

# 9. Input Specification

### Request DTO (JSON)

```json
{
  "productItemId": "c20befe6-e4b9-4f37-b514-e66233ef04a1",
  "theme": "CLASSIC_NAVY",
  "includeAuditLogs": false
}
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| productItemId | UUID | Yes | Phải tồn tại trong `product_items` | ID sản phẩm cần xuất tài liệu hành trình |
| theme | String | No | Enum: `CLASSIC_NAVY`, `WARM_MINIMAL` | Chủ đề màu sắc giao diện PDF |
| includeAuditLogs| Boolean | No | Mặc định: `false` | Có gộp thêm nhật ký chỉnh sửa hệ thống vào phụ lục |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "message": "PDF export job initiated.",
  "data": {
    "jobId": "993fa1a0-1200-4b2e-a551-fb112aaee111",
    "status": "PROCESSING",
    "estimatedTimeSeconds": 5
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `events`: SELECT lấy mốc sự kiện.
- `product_items`: SELECT lấy thông tin sản phẩm.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `events` | SELECT | Lấy thông tin lịch sử phục vụ in ấn |
| `audit_logs` | INSERT | Nhật ký hóa hành vi export |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-XPD-001 | Link tải file PDF (Signed URL) sinh ra bắt buộc phải có thời gian hết hạn (ví dụ: chỉ có hiệu lực tải trong vòng 24 giờ kể từ lúc tạo) để đảm bảo an ninh thông tin, tránh lộ link ra ngoài. |
| BR-XPD-002 | Tệp PDF phải có phần chữ ký điện tử số của hệ thống (Digital watermarking/hash checksum) in ở góc trang để tránh việc người ngoài tự chỉnh sửa file PDF giả mạo thông tin. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| productItemId | required, valid UUID | "ID sản phẩm đơn lẻ không đúng định dạng" |

---

# 14. API Design

### Endpoint

- **Method**: POST
- **Path**: `/api/v1/trace/export/pdf`

### HTTP Status

| Status | Meaning |
|---|---|
| 202 | Accepted. Tiến trình sinh PDF ngầm đã được khởi chạy |
| 400 | Bad Request. ID sản phẩm trống |
| 401 | Unauthorized. Chưa đăng nhập |
| 404 | Not Found. Không tìm thấy thông tin sản phẩm |
| 500 | Internal Server Error. Lỗi dựng file hoặc lỗi thư viện render |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình Timeline sản phẩm (`/trace/:id`).
- **Component**: Nút bấm `ExportPDFButton` - Khi click sẽ hiển thị trạng thái xoay vòng (Spinner) kèm text "Generating PDF document..." cho đến khi nhận được link tải hoặc nhận được thông báo đẩy.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: POST `/api/v1/trace/export/pdf`
- **Responsibility**: Validate token, validate input body, gọi `TraceExportService.initiatePdfExport` để đẩy vào queue ngầm.

### Service
- **TraceExportService**:
  1. Đọc dữ liệu từ Postgres.
  2. Tạo job gửi thông tin sang Broker RabbitMQ.
  3. Trả về mã 202 cho Controller phản hồi khách hàng.
  4. (Worker ngầm nhận task) Sử dụng công cụ render HTML sang PDF (Ví dụ: mẫu EJS template kết hợp thư viện `pdfkit` / `puppeteer` / `html-pdf`):
     - Vẽ đầu trang có Logo doanh nghiệp.
     - Dựng bảng thông tin chung: Số lô, SKU, ngày hết hạn.
     - Vẽ chuỗi timeline xương cá tinh tế thể hiện từng mốc di chuyển.
     - Đóng gói xuất file ra đĩa.
  5. Upload file lên Cloud Storage.
  6. Publish event `trace.exported` sang RabbitMQ để thông báo cho khách hàng.

### Event Payload (RabbitMQ)

- **Exchange**: `notification.events`
- **Routing Key**: `trace.exported`
- **Payload**:
```json
{
  "eventId": "zz3cc812-7bb1-4110-8aa2-9f881b2a99zz",
  "eventType": "trace.exported",
  "timestamp": "2026-07-08T18:06:00Z",
  "data": {
    "userId": "a20befe6-e4b9-4f37-b514-e66233ef04a1",
    "format": "PDF",
    "downloadUrl": "https://producttrace.ai/storage/temp/export-993fa.pdf",
    "fileName": "ProductJourney-SN0001.pdf"
  }
}
```

---

# 17. Event Flow

```
[User] -> Click Export -> [TraceExportService] -> Send Job -> [RabbitMQ Broker]
                                                                   |
                                                                   +---> Consumer: [Worker Process]
                                                                                      |
                                                                                      +---> (PostgreSQL Query)
                                                                                      +---> (Generate PDF)
                                                                                      +---> (Upload Cloud Storage)
                                                                                      +---> Publish trace.exported -> [RabbitMQ]
                                                                                                                           |
                                                                                                                           +---> Consumer: [Notification Service] -> (Send In-App / Email Alert)
```

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| EXPORT_TRACE_TIMELINE_PDF | Admin / Staff / Dealer | `userId`, `productItemId`, `downloadUrl`, `timestamp` |

---

# 19. Security Consideration

- **Signed URL Expiration**: Bắt buộc link tải file PDF lưu trong S3/GCS phải tự động hết hiệu lực sau tối đa 24 giờ.
- **Resource Protection**: Giới hạn tối đa 3 lượt yêu cầu xuất PDF trên 5 phút đối với một tài khoản để ngăn chặn hacker viết script gọi liên tiếp nhằm ddos hạ gục máy chủ render.

---

# 20. Acceptance Criteria

### Scenario: Xuất in tài liệu PDF hành trình sữa tươi thành công
- **Given** sản phẩm có mốc vận chuyển đầy đủ và tài khoản có quyền `EXPORT_TRACE_PDF`.
- **When** nhân viên nhấn "Export PDF".
- **Then** hệ thống khởi chạy job ngầm, xuất file PDF đẹp đẽ chứa đầy đủ 2 sự kiện kèm chữ ký checksum của hệ thống dưới chân trang, lưu lên lưu trữ đám mây an toàn và gửi thông báo thông báo cho người dùng kèm nút "Download File".

---

# 21. Developer Checklist

### Backend
- [ ] Thiết lập API POST `/api/v1/trace/events/export-pdf`.
- [ ] Tích hợp công cụ render tài liệu PDF (ví dụ: `pdfkit` / `puppeteer` / `playwright`).
- [ ] Viết hàm lưu trữ file PDF vào Storage (Local/Cloud Run storage).
- [ ] Tích hợp RabbitMQ publish event `trace.exported` để báo cho Notification Service.

### Frontend
- [ ] Thiết kế nút "Xuất PDF" kèm icon máy in đẹp đẽ.
- [ ] Xử lý trigger tải xuống (Download trigger) khi nhận được socket thông báo từ Notification service.
- [ ] Thêm thông báo pop-up đẹp mắt báo hiệu tiến trình đang xuất ngầm.
