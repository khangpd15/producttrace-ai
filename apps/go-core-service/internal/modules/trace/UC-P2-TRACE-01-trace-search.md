# UC-P2-TRACE-01 - Tìm kiếm timeline

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Traceability |
| Use Case ID | UC-P2-TRACE-01 |
| Feature Name | Tìm kiếm timeline (Trace Search API) |
| Priority | Medium |
| Git Branch | feature/trace-search |
| Producer | - |
| Consumer | Frontend |
| Dependency | UC-P1-TRACE-03 (View Event List) |

## Description

- **Business Purpose**: Cung cấp khả năng tìm kiếm nhanh hành trình lịch sử (Timeline) của một sản phẩm cụ thể bằng cách nhập mã sản phẩm (`item_code` / mã QR) hoặc số Serial, giúp giải đáp tức thì câu hỏi "sản phẩm này có nguồn gốc từ đâu và đã đi qua các chặng nào".
- **User Problem Solved**: Khách hàng hoặc cán bộ quản lý chất lượng có một sản phẩm thực tế và muốn biết toàn bộ hành trình của nó nhưng không có link quét trực tiếp. Họ cần một ô nhập để tra cứu thủ công nhanh chóng.
- **Expected System Behavior**: Hệ thống tiếp nhận mã định danh từ người dùng, thực hiện truy vấn bảng `product_items` để lấy thông tin sản phẩm mẹ và thông tin Batch liên quan. Tiếp theo, hệ thống tìm toàn bộ các bản ghi trong bảng `events` liên kết với `product_item_id` này hoặc `batch_id` của nó, sắp xếp theo thời gian và trả về Frontend hiển thị dạng dòng lịch sử dọc.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Customer | Nhập mã trên trang chủ (Public portal) để kiểm tra hàng chính hãng và xem hành trình sản phẩm. |
| Staff Kho / Dealer| Tra cứu thông tin di chuyển sản phẩm để phục vụ kiểm kho hoặc xử lý bảo hành. |
| Admin | Quản trị viên kiểm tra dữ liệu chuỗi cung ứng khi có khiếu nại từ khách hàng. |

---

# 3. Permission

| Role | Create | Update | Delete | View |
|---|---|---|---|---|
| Admin | No | No | No | Yes |
| Staff Kho | No | No | No | Yes |
| Dealer | No | No | No | Yes |
| Customer | No | No | No | Yes |

---

# 4. Preconditions

- Dữ liệu sản phẩm đơn lẻ (`ProductItem`) đã tồn tại trong database Postgres.
- Sản phẩm đã được gán nhãn QR hoặc có số Serial hoạt động.

---

# 5. Trigger

User truy cập trang tra cứu công cộng (Public Trace portal) hoặc màn hình quản lý, nhập mã vào ô "Tra cứu nguồn gốc" và click nút "Search".

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Nhập mã sản phẩm (ví dụ: `PT-MILK-SN0001`) hoặc số Serial vào ô tìm kiếm và bấm Enter. | Gửi request GET tới API `/api/v1/trace/search` kèm query parameter `code=PT-MILK-SN0001`. |
| 2 | System | Chuẩn hóa đầu vào | Thực hiện trim các dấu cách thừa và chuyển chuỗi ký tự thành dạng chữ in hoa. |
| 3 | System | Truy vấn Sản phẩm | Thực hiện SELECT bảng `product_items` WHERE `item_code = :code` hoặc `serial_number = :code`. Nếu không tìm thấy, trả lỗi 404. |
| 4 | System | Truy vấn Timeline | Lấy toàn bộ danh sách `events` liên kết với `product_item_id` này và gộp cả các sự kiện cấp độ `batch_id` liên quan. |
| 5 | System | Sắp xếp dữ liệu | Sắp xếp mảng sự kiện theo thời gian tăng dần (`created_at ASC`) để hiển thị theo đúng thứ tự logic hành trình di chuyển thực tế từ xưa đến nay. |
| 6 | System | Ghi Audit Log | Lưu lại hành động tra cứu nguồn gốc của User vào log. |
| 7 | System | Trả kết quả | Phản hồi thông tin sản phẩm và mảng timeline hoàn chỉnh dạng JSON về cho Frontend. |
| 8 | User | Xem Timeline | Giao diện vẽ biểu đồ dòng thời gian (Life-cycle timeline) cực kỳ bắt mắt với các mốc hoạt họa chi tiết. |

---

# 7. Alternative Flow

## AF-001 Tra cứu bằng Camera quét QR trực tiếp (Webcam QR Scan)
- **Description**: Khách hàng bấm vào biểu tượng Camera trên thanh tìm kiếm, cấp quyền camera cho trình duyệt để quét trực tiếp mã QR in trên hộp sản phẩm. Sau khi nhận diện được chuỗi URL, hệ thống tự động tách lấy `item_code` ở cuối và tự động gọi API tìm kiếm timeline.

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Mã sản phẩm không tồn tại trong hệ thống | Trả về HTTP 404: "Product item not found. Please verify your product code and try again." |
| ERR-002 | Sản phẩm bị khóa do lô sản xuất tương ứng bị thu hồi | Hệ thống vẫn hiển thị timeline nhưng có Banner đỏ cảnh báo: "WARNING: This product belongs to a recalled batch. Please do not consume." |

---

# 9. Input Specification

### Request Query Parameters

```
GET /api/v1/trace/search?code=PT-MILK-SN0001
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| code | String | Yes | Dài từ 3 đến 100 ký tự | Mã QR `item_code` hoặc `serial_number` |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "data": {
    "productItem": {
      "itemId": "c20befe6-e4b9-4f37-b514-e66233ef04a1",
      "itemCode": "PT-MILK-SN0001",
      "serialNumber": "SN-2026-0001",
      "status": "AVAILABLE",
      "productName": "Sữa tươi tiệt trùng Vinamilk 1L",
      "thumbnailUrl": "https://producttrace.ai/images/milk-1l.jpg"
    },
    "timeline": [
      {
        "eventId": "bb3fc812-7bb1-4110-8aa2-9f881b2a99aa",
        "eventType": "PRODUCED",
        "title": "Sản xuất thành công",
        "description": "Sản phẩm được chế biến và đóng chai tại Nhà máy sữa Bình Dương",
        "location": "Bình Dương, Việt Nam",
        "timestamp": "2026-07-01T08:00:00Z"
      },
      {
        "eventId": "cc3fc812-7bb1-4110-8aa2-9f881b2a99bb",
        "eventType": "WAREHOUSE_IN",
        "title": "Nhập kho trung chuyển Sài Gòn",
        "description": "Bảo quản lạnh tiêu chuẩn",
        "location": "Quận 9, TP. Hồ Chí Minh",
        "timestamp": "2026-07-03T14:30:00Z"
      }
    ]
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `product_items`: SELECT thông tin hộp sản phẩm.
- `batches`: SELECT thông tin hạn sử dụng và lô liên quan.
- `events`: SELECT danh sách sự kiện di chuyển.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `product_items`| SELECT | Kiểm tra sự tồn tại và lấy ID nội bộ của sản phẩm |
| `events` | SELECT | Lấy chuỗi lịch sử của sản phẩm và của cả lô liên quan |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-TRS-001 | Trang tra cứu nguồn gốc của sản phẩm đơn lẻ phải công khai, không yêu cầu đăng nhập đối với Khách hàng quét mã bên ngoài. |
| BR-TRS-002 | Timeline phải hiển thị các sự kiện theo thứ tự tăng dần của thời gian (`created_at ASC`) để đảm bảo trải nghiệm đọc hành trình từ lúc sinh ra đến lúc tiêu dùng. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| code | required, min_length 3 | "Vui lòng nhập mã sản phẩm hoặc số Serial hợp lệ để tìm kiếm" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/trace/search`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Tìm thấy sản phẩm và trả về timeline chi tiết |
| 400 | Bad Request. Mã tìm kiếm rỗng |
| 404 | Not Found. Không tìm thấy mã sản phẩm này trong database |
| 500 | Internal Server Error. Lỗi database |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình công cộng "Trace Product Journey" (`/trace`).
- **Components**:
  - `TimelineSearchForm`: Ô nhập mã to, đặt ở vị trí trung tâm, có nút kích hoạt camera quét QR Code.
  - `TraceTimelineVisualizer`: Khung vẽ timeline nghệ thuật. Mỗi nút mốc thời gian có một biểu tượng tương ứng (nhà máy cho `PRODUCED`, xe tải cho `IN_TRANSIT`, giỏ hàng cho `SALE`).

### UI State
- **No Permissions Banner**: Nếu có các sự kiện nội bộ của kho bãi (nhập/xuất nội bộ) mà hệ thống cấu hình ẩn với khách, chỉ hiển thị các mốc phân phối chính để bảo vệ bí mật kinh doanh.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/trace/search`
- **Responsibility**: Nhận mã từ query string, gọi `TraceQueryService.searchTimeline`.

### Service
- **TraceQueryService**:
  1. Truy vấn PostgreSQL tìm dòng dữ liệu trong `product_items` bằng `item_code` hoặc `serial_number`.
  2. Nếu không có, ném lỗi `ProductItemNotFoundException`.
  3. Query danh sách `events` có `product_item_id = item.id` HOẶC `batch_id = item.batch_id`.
  4. Sắp xếp mảng gộp này theo ngày tạo tăng dần.
  5. Đóng gói dữ liệu trả về cho Controller.

---

# 17. Event Flow

*Không phát sinh RabbitMQ Event mới vì đây là luồng đọc dữ liệu công khai (Read-only API).*

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| PUBLIC_SEARCH_TIMELINE | Public / Customer | `searchedCode`, `ipAddress`, `userAgent`, `timestamp` |

---

# 19. Security Consideration

- **Brute Force Defense**: Sử dụng rate limiting (giới hạn 30 lượt search/phút trên 1 IP) để chống các con bot cào dữ liệu (scraping bots) cố tình dò quét dải số Serial của doanh nghiệp.

---

# 20. Acceptance Criteria

### Scenario: Khách hàng tra cứu mã QR hợp lệ thành công
- **Given** sản phẩm sữa chua có mã QR `PT-YOGURT-998` đã được bán cho khách hàng.
- **When** khách hàng gõ mã `PT-YOGURT-998` vào trang chủ tra cứu.
- **Then** hệ thống hiển thị chính xác tên sản phẩm là "Sữa chua Vinamilk nếp cẩm", kèm biểu đồ timeline dọc gồm 3 mốc: "Sản xuất tại Bình Dương", "Nhập kho Hà Nội", và "Bán tại siêu thị Co.opmart".

---

# 21. Developer Checklist

### Backend
- [ ] Thiết lập API GET `/api/v1/trace/search` không yêu cầu JWT auth.
- [ ] Viết truy vấn lấy gộp cả sự kiện của product_item và sự kiện của lô hàng cha.
- [ ] Implement Rate Limiter bằng Redis hoặc middleware in-memory.

### Frontend
- [ ] Xây dựng trang tra cứu nguồn gốc phong cách tối giản, sang trọng (Minimalist, warm theme).
- [ ] Tích hợp thư viện quét QR bằng Camera (ví dụ: `html5-qrcode` hoặc `react-qr-reader`).
- [ ] Thêm hiệu ứng transition mượt mà bằng `motion/react` khi hiển thị các mốc timeline.
