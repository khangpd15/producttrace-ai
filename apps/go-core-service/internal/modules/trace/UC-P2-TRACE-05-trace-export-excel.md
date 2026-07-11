# UC-P2-TRACE-05 - Xuất Excel lịch sử truy vết

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Traceability |
| Use Case ID | UC-P2-TRACE-05 |
| Feature Name | Xuất Excel lịch sử truy vết (Excel Export API) |
| Priority | Medium |
| Git Branch | feature/trace-export-excel |
| Producer | trace.exported |
| Consumer | Notification |
| Dependency | UC-P1-TRACE-03 (View Event List) |

## Description

- **Business Purpose**: Cung cấp khả năng kết xuất toàn bộ danh sách các mốc thời gian, dữ liệu siêu dữ liệu (Metadata) và thông tin sản phẩm của một hay nhiều sản phẩm đơn lẻ ra định dạng file bảng tính Microsoft Excel (.xlsx). Việc này phục vụ đắc lực cho công tác kiểm duyệt, báo cáo dữ liệu định kỳ cho ban giám đốc hoặc gửi số liệu phân tích sang hệ thống ERP/kế toán ngoài.
- **User Problem Solved**: Ban giám đốc yêu cầu báo cáo thống kê tình hình luân chuyển hàng hóa tuần qua dưới dạng số liệu thô (raw data) để tính toán KPI vận chuyển của các đối tác giao hàng. Nhân viên không thể ngồi copy từng dòng trên màn hình web. Họ cần một file Excel hoàn chỉnh chứa đầy đủ bảng biểu thô để tiện tính toán bằng các hàm `SUM`, `AVERAGE` hoặc vẽ biểu đồ PivotTable.
- **Expected System Behavior**: Hệ thống nhận yêu cầu xuất, truy vấn cơ sở dữ liệu PostgreSQL để gom thông tin hành trình sự kiện, sử dụng thư viện xử lý file Excel (Ví dụ: `exceljs` hoặc `xlsx`), xây dựng một file bảng tính có định dạng đẹp mắt (có tô màu header, tự động co giãn độ rộng cột - auto-fit column width), upload lên Cloud Storage, phát đi RabbitMQ event `trace.exported` loại `EXCEL` để Notification Service gửi link tải trực tiếp đến cho User.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Admin / Supervisor | Xuất file dữ liệu để thực hiện đối toán thời gian luân chuyển của đối tác vận chuyển. |
| Staff Kho | Xuất báo cáo hoạt động nhập/xuất kho trong tháng để làm biên bản kiểm kê định kỳ. |
| Notification Service| Consumer nhận event để gửi Email hoặc In-app notification thông báo file Excel đã tạo xong. |

---

# 3. Permission

| Role | Create | Update | Delete | View |
|---|---|---|---|---|
| Admin | Yes | No | No | Yes |
| Staff Kho | Yes | No | No | Yes |
| Dealer | Yes | No | No | Yes |
| Customer | No | No | No | No |

---

# 4. Preconditions

- Người dùng đã đăng nhập và được cấp quyền `EXPORT_TRACE_EXCEL`.
- Có dữ liệu sản phẩm và sự kiện liên quan lưu trong PostgreSQL.

---

# 5. Trigger

User click vào nút "Export Excel" (Biểu tượng bảng tính màu xanh lá) tại màn hình Timeline sản phẩm hoặc trang quản lý báo cáo Traceability.

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại trang chi tiết hành trình, nhấn nút "Export Excel" và chọn khoảng thời gian cần kết xuất. | Hiển thị hộp thoại cấu hình xuất và gửi request POST sang API. |
| 2 | System | Tiếp nhận Yêu cầu | Đăng ký một Job xử lý ngầm (Asynchronous worker task) để tránh treo luồng xử lý chính. Trả về mã HTTP 202 Accepted. |
| 3 | System | Query Database | Worker ngầm truy vấn PostgreSQL lấy thông tin `product_items`, `products`, `batches`, `locations` và mảng `events`. |
| 4 | System | Tạo bảng tính | Khởi tạo workbook Excel mới bằng thư viện `exceljs`. Thiết lập 2 Worksheet: Sheet 1 chứa "Thông tin sản phẩm chung", Sheet 2 chứa "Nhật ký hành trình truy vết chi tiết". |
| 5 | System | Định dạng Excel | Trang trí file Excel chuyên nghiệp: tô màu xanh lá cây đậm cho header, định dạng cột ngày tháng chuẩn `YYYY-MM-DD HH:mm:ss`, đặt cờ tự động giãn cột dựa trên độ dài chữ. |
| 6 | System | Lưu trữ File | Ghi file ra ổ cứng tạm, upload lên Cloud Storage và sinh Signed URL tải file an toàn có thời hạn (Hết hạn sau 24 giờ). |
| 7 | System | Gửi Event RabbitMQ | Publish tin nhắn `trace.exported` loại `EXCEL` lên RabbitMQ exchange `trace.exports`. |
| 8 | System | Ghi Audit Log | Ghi lại hành vi xuất Excel thành công của User. |
| 9 | System | Trả thông báo | Notification Service nhận event, đẩy thông báo thời gian thực (Websocket) hoặc gửi Email kèm nút tải trực tiếp cho người dùng. |

---

# 7. Alternative Flow

## AF-001 Xuất dữ liệu hàng loạt cho cả lô sản xuất (Batch-level Bulk Excel Export)
- **Description**: Cho phép Admin chọn một Batch ID và xuất Excel chứa hành trình của toàn bộ 1000 sản phẩm thuộc lô hàng đó gộp chung vào một file Excel duy nhất (Mỗi sản phẩm nằm trên một hàng, hoặc mỗi sản phẩm là một Worksheet riêng lẻ).

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Số lượng bản ghi cần xuất quá lớn (> 50,000 dòng sự kiện) | Hệ thống từ chối xuất đồng bộ, yêu cầu cấu hình thêm bộ lọc thời gian để thu hẹp dải dữ liệu: "Dataset too large. Please narrow your date range filter." |
| ERR-002 | Lỗi mất kết nối hàng đợi RabbitMQ | Worker ngầm vẫn cố gắng hoàn thành tệp, lưu link tải trực tiếp vào lịch sử xuất báo cáo của User trong DB và trả mã 500 nếu mất kết nối hoàn toàn. |

---

# 9. Input Specification

### Request DTO (JSON)

```json
{
  "productItemId": "c20befe6-e4b9-4f37-b514-e66233ef04a1",
  "batchId": "6f2bc881-8b21-4f10-9111-a887b2210a12",
  "fromDate": "2026-07-01T00:00:00Z",
  "toDate": "2026-07-08T23:59:59Z"
}
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| productItemId | UUID | No | Phải tồn tại trong `product_items` | ID sản phẩm đơn lẻ (nếu xuất lẻ) |
| batchId | UUID | No | Phải tồn tại trong `batches` | ID lô hàng (nếu xuất cả lô) |
| fromDate | DateTime | No | ISO 8601 | Bộ lọc mốc thời gian từ ngày |
| toDate | DateTime | No | ISO 8601 | Bộ lọc mốc thời gian đến ngày |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "message": "Excel export job initiated.",
  "data": {
    "jobId": "ee8fa1a0-1200-4b2e-a551-fb112aaee222",
    "status": "PROCESSING",
    "estimatedTimeSeconds": 3
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `events`: SELECT lấy lịch sử.
- `product_items`: SELECT thông tin.
- `batches`: SELECT thông tin lô hàng.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `events` | SELECT | Lấy thông tin các mốc sự kiện thô |
| `audit_logs` | INSERT | Ghi nhận hành vi export Excel |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-XLS-001 | Bố cục cột trong bảng Excel phải cố định và thống nhất: `STT`, `Event ID`, `Loại sự kiện`, `Tiêu đề`, `Mô tả`, `Địa điểm`, `Người thực hiện`, `Thời gian (UTC)`. |
| BR-XLS-002 | Link tải file Excel sinh ra phải được bảo mật bằng cơ chế Token có thời hạn hết hạn tự động (Expired after 24 hours) để bảo vệ dữ liệu nội bộ chuỗi cung ứng. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| productItemId | UUID format | "Mã ID sản phẩm không hợp lệ" |

---

# 14. API Design

### Endpoint

- **Method**: POST
- **Path**: `/api/v1/trace/export/excel`

### HTTP Status

| Status | Meaning |
|---|---|
| 202 | Accepted. Tiến trình khởi chạy thành công |
| 400 | Bad Request. Sai cấu trúc tham số đầu vào |
| 401 | Unauthorized. Chưa đăng nhập |
| 403 | Forbidden. Không có quyền xuất báo cáo Excel |
| 500 | Internal Server Error. Gặp lỗi khi render tệp hoặc lưu trữ đám mây |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình xem hành trình sản phẩm hoặc trang Báo cáo thống kê.
- **Component**: Nút bấm `ExportExcelButton` - Có icon Microsoft Excel màu xanh lá cây, khi click sẽ có hiệu ứng xoay tròn và hiển thị tooltip thông báo.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: POST `/api/v1/trace/export/excel`
- **Responsibility**: Nhận request, check token, gọi `TraceExportService.initiateExcelExport`.

### Service
- **TraceExportService**:
  1. Đọc dữ liệu thô từ cơ sở dữ liệu.
  2. Tạo tin nhắn job chuyển giao cho RabbitMQ.
  3. Trả về mã 202 ngay tức khắc để giải phóng thread.
  4. (Worker ngầm nhận task) Sử dụng thư viện `exceljs` để xây dựng bảng:
     - Tạo workbook.
     - Tạo style định dạng phông chữ Inter cỡ 11, tô đậm hàng đầu, tô nền xám nhạt cho hàng chẵn để dễ đọc (Zebra striping).
     - Đẩy mảng dữ liệu vào worksheet.
     - Ghi tệp ra ổ đĩa đám mây.
  5. Publish event `trace.exported` sang RabbitMQ để kích hoạt gửi thông báo cho khách hàng.

### Event Payload (RabbitMQ)

- **Exchange**: `notification.events`
- **Routing Key**: `trace.exported`
- **Payload**:
```json
{
  "eventId": "xx3cc812-7bb1-4110-8aa2-9f881b2a99xx",
  "eventType": "trace.exported",
  "timestamp": "2026-07-08T18:06:00Z",
  "data": {
    "userId": "a20befe6-e4b9-4f37-b514-e66233ef04a1",
    "format": "EXCEL",
    "downloadUrl": "https://producttrace.ai/storage/temp/export-ee8fa.xlsx",
    "fileName": "ProductJourney-SN0001.xlsx"
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
                                                                                      +---> (Generate Excel Workbook with styles)
                                                                                      +---> (Upload Cloud Storage)
                                                                                      +---> Publish trace.exported -> [RabbitMQ]
                                                                                                                           |
                                                                                                                           +---> Consumer: [Notification Service] -> (Email/In-App Link)
```

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| EXPORT_TRACE_TIMELINE_EXCEL | Admin / Staff / Dealer | `userId`, `productItemId`, `downloadUrl`, `timestamp` |

---

# 19. Security Consideration

- **Sensitive Columns Filtering**: Chặn đứng tuyệt đối không được ghi các dữ liệu nhạy cảm của hệ thống như mật khẩu hash, token thanh toán hay mã bảo mật riêng tư vào trong file Excel tải về.

---

# 20. Acceptance Criteria

### Scenario: Xuất dữ liệu Excel hành trình sữa tươi thành công
- **Given** người dùng là Admin của tổng kho Bình Dương, đang xem hành trình của sản phẩm `MILK-123`.
- **When** người dùng nhấn nút "Export Excel".
- **Then** hệ thống khởi tạo tác vụ xuất ngầm, tạo file Excel có 2 sheet định dạng chuyên nghiệp chuẩn chỉ, lưu lên kho lưu trữ và thông báo cho người dùng link tải trực tiếp trong vòng 5 giây.

---

# 21. Developer Checklist

### Backend
- [ ] Thiết lập API POST `/api/v1/trace/events/export-excel`.
- [ ] Cấu hình thư viện `exceljs` để xây dựng workbook và định dạng style chữ/bảng.
- [ ] Viết hàm upload file lên Cloud Storage.
- [ ] Gửi RabbitMQ message `trace.exported` loại EXCEL để kích hoạt gửi thông báo.

### Frontend
- [ ] Thiết kế nút bấm "Xuất Excel" màu xanh lá cây mát mắt.
- [ ] Lắng nghe socket sự kiện từ server để tự động kích hoạt tiến trình tải xuống (Auto-download) khi file đã sẵn sàng.
- [ ] Xử lý lỗi đẹp đẽ nếu bị nghẽn mạng.
