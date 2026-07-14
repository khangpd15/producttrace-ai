# UC-P2-TRACE-02 - Lọc timeline theo thời gian

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Traceability |
| Use Case ID | UC-P2-TRACE-02 |
| Feature Name | Lọc timeline theo thời gian (Trace Filter API) |
| Priority | Medium |
| Git Branch | feature/trace-filter-date |
| Producer | - |
| Consumer | Frontend |
| Dependency | UC-P2-TRACE-01 (Trace Search) |

## Description

- **Business Purpose**: Cho phép người quản trị, nhân viên kiểm kho lọc và khoanh vùng các sự kiện lịch sử (Timeline) xảy ra trong một khoảng thời gian xác định (Từ ngày ... Đến ngày ...). Việc này hỗ trợ cho công tác kiểm duyệt tiến độ vận chuyển hàng tuần hoặc kiểm tra chuỗi lạnh (Cold chain verification) trong các ngày nắng nóng đỉnh điểm.
- **User Problem Solved**: Khi sản phẩm có lịch sử phân phối kéo dài nhiều năm hoặc đi qua hàng chục khâu trung chuyển trung gian, timeline sẽ trở nên quá dài và khó đọc. Người dùng muốn lọc nhanh xem trong tuần qua có những sự kiện gì đã xảy ra với sản phẩm này để làm báo cáo.
- **Expected System Behavior**: Hệ thống tiếp nhận các tham số ngày bắt đầu (`fromDate`) và ngày kết thúc (`toDate`), chuyển đổi thành mốc thời gian UTC, thực hiện truy vấn SELECT trong PostgreSQL với mệnh đề `created_at BETWEEN :fromDate AND :toDate`, rồi trả về mảng sự kiện đã được lọc cho người dùng.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Staff Kho | Lọc sự kiện theo ngày để kiểm tra lịch nhập kho thực tế của đợt sữa trong tuần. |
| Admin / Supervisor | Kiểm duyệt các sự kiện vận chuyển và lưu kho theo khung thời gian kiểm toán. |
| Dealer / Store | Kiểm tra các mốc phân phối của hàng hóa trong tháng hiện tại. |

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

- Người dùng đã tìm kiếm được mã sản phẩm hoặc đang ở trong trang chi tiết Timeline của sản phẩm (`UC-P2-TRACE-01`).
- Mốc thời gian lọc từ ngày và đến ngày phải hợp lệ (Ví dụ: Từ ngày <= Đến ngày).

---

# 5. Trigger

User chọn khoảng ngày trên bộ lọc lịch (Date Range Picker) tại trang hiển thị Timeline của sản phẩm và nhấn nút "Apply Filter".

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại trang Timeline sản phẩm, mở bộ chọn khoảng ngày, chọn ngày bắt đầu (ví dụ: `2026-07-01`) và ngày kết thúc (ví dụ: `2026-07-07`), rồi nhấn nút "Apply". | Gọi API `/api/v1/trace/search` hoặc API phụ trợ lọc kèm query params `fromDate=2026-07-01T00:00:00Z&toDate=2026-07-07T23:59:59Z`. |
| 2 | System | Validate Date Format| Kiểm tra định dạng ngày gửi lên (ISO 8601), xác minh điều kiện `fromDate <= toDate`. |
| 3 | System | Query Database | Thực hiện truy vấn SELECT bảng `events` kết hợp mệnh đề logic `AND e.created_at >= :fromDate AND e.created_at <= :toDate`. |
| 4 | System | Sắp xếp mốc | Sắp xếp dữ liệu theo ngày tăng dần/giảm dần tùy theo cấu hình của người dùng. |
| 5 | System | Ghi Audit Log | Nhật ký hóa hành động lọc lịch sử theo thời gian của người dùng. |
| 6 | System | Trả kết quả | Phản hồi danh sách các sự kiện nằm trong khoảng thời gian đã chọn về cho Frontend. |
| 7 | User | Đọc thông tin | Giao diện thu hẹp và vẽ lại timeline, chỉ hiển thị các mốc thời gian khớp với khoảng ngày được chọn. |

---

# 7. Alternative Flow

*Không có luồng thay thế đặc thù.*

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Ngày bắt đầu lớn hơn ngày kết thúc | Trả về HTTP 400 Bad Request: "Invalid date range. Start date cannot be after end date." |
| ERR-002 | Ngày gửi lên sai định dạng (Ví dụ: `31/02/2026`) | Trả về lỗi: "Invalid query parameters. Date must follow ISO-8601 format." |

---

# 9. Input Specification

### Request Query Parameters

```
GET /api/v1/trace/search?code=PT-MILK-SN0001&fromDate=2026-07-01T00:00:00Z&toDate=2026-07-07T23:59:59Z
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| fromDate | DateTime | No | ISO 8601, <= `toDate` | Lọc các sự kiện bắt đầu từ mốc giờ này |
| toDate | DateTime | No | ISO 8601, >= `fromDate` | Lọc các sự kiện kết thúc ở mốc giờ này |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "data": {
    "filterApplied": {
      "fromDate": "2026-07-01T00:00:00Z",
      "toDate": "2026-07-07T23:59:59Z"
    },
    "matchedEventsCount": 1,
    "timeline": [
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
- `events`: SELECT các sự kiện thỏa mãn điều kiện thời gian.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `events` | SELECT | Lọc danh sách mốc thời gian trong khoảng `created_at` |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-FDT-001 | Nếu không truyền `fromDate` và `toDate`, hệ thống mặc định hiển thị toàn bộ lịch sử không giới hạn thời gian (No limit). |
| BR-FDT-002 | Thời gian nhập vào của người dùng theo múi giờ địa phương (local timezone) phải được chuyển đổi đồng bộ sang mốc giờ UTC ở Server trước khi thực hiện so sánh với database PostgreSQL. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| fromDate | is_valid_date | "Ngày bắt đầu không đúng định dạng thời gian" |
| toDate | is_valid_date | "Ngày kết thúc không đúng định dạng thời gian" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/trace/search`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Trả về danh sách đã lọc thành công (Có thể trả về danh sách rỗng) |
| 400 | Bad Request. Khoảng ngày bị ngược hoặc sai format |
| 404 | Not Found. Không tìm thấy mã sản phẩm ban đầu |
| 500 | Internal Server Error. Gặp lỗi hệ thống |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình xem hành trình sản phẩm.
- **Component**: `DateRangePicker` - Thanh lịch chọn khoảng ngày (như thư viện `react-day-picker`), cho phép người dùng nhấp đúp chọn ngày nhanh, hiển thị nhãn "7 ngày qua", "30 ngày qua" để thao tác nhanh.

### UI State
- **Reset Button**: Nút "Xóa bộ lọc" (Clear Filter button) xuất hiện kế bên để người dùng hủy lọc thời gian chỉ bằng 1 lượt nhấp chuột.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/trace/search`
- **Responsibility**: Parse các query `fromDate` và `toDate`, gọi `TraceQueryService.searchTimeline` có kèm bộ lọc thời gian.

### Service
- **TraceQueryService**:
  1. Kiểm tra mã sản phẩm.
  2. Xây dựng truy vấn SQL động:
     ```sql
     SELECT * FROM events
     WHERE product_item_id = :itemId
       AND (:fromDate IS NULL OR created_at >= :fromDate)
       AND (:toDate IS NULL OR created_at <= :toDate)
     ORDER BY created_at ASC
     ```
  3. Trả về mảng timeline cho Controller.

---

# 17. Event Flow

```
[User] -> Select Dates -> [TraceQueryService]
                                |
                                +---> (PostgreSQL Query: WHERE created_at BETWEEN fromDate AND toDate)
```

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| FILTER_TIMELINE_BY_DATE | Staff / Customer | `userId`, `searchedCode`, `fromDate`, `toDate`, `timestamp` |

---

# 19. Security Consideration

- **SQL Injection Prevention**: Thực hiện validate bắt buộc kiểu dữ liệu đầu vào là dạng Date để chặn đứng hoàn toàn việc hacker gài các chuỗi mã SQL độc hại vào tham số lọc ngày tháng.

---

# 20. Acceptance Criteria

### Scenario: Lọc sự kiện trong tuần đầu tiên của tháng 7
- **Given** sản phẩm có 5 sự kiện trong tháng 7/2026. Có 1 sự kiện xảy ra ngày `2026-07-03` và 4 sự kiện còn lại xảy ra trong tháng 8.
- **When** người dùng cấu hình lọc ngày từ `2026-07-01` đến `2026-07-07`.
- **Then** hệ thống gọi API lọc và trả về đúng 1 sự kiện ngày `2026-07-03`, ẩn các sự kiện của tháng 8 đi.

---

# 21. Developer Checklist

### Backend
- [ ] Bổ sung query parameters `fromDate`, `toDate` cho API tra cứu timeline.
- [ ] Viết hàm chuyển đổi timezone (ví dụ: dùng thư viện `moment` hoặc `date-fns` ở backend).
- [ ] Thêm điều kiện truy vấn động trong ORM/SQL query.

### Frontend
- [ ] Xây dựng giao diện thanh Date Range Picker thân thiện.
- [ ] Đồng bộ trạng thái lịch chọn với URL query params.
- [ ] Hiển thị thông báo "Không có sự kiện nào xảy ra trong khoảng thời gian này" nếu danh sách rỗng.
