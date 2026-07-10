# UC-P2-TRACE-03 - Lọc timeline theo loại sự kiện

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Traceability |
| Use Case ID | UC-P2-TRACE-03 |
| Feature Name | Lọc timeline theo loại sự kiện (Trace Event Filter API) |
| Priority | Medium |
| Git Branch | feature/trace-filter-event |
| Producer | - |
| Consumer | Frontend |
| Dependency | UC-P2-TRACE-01 (Trace Search) |

## Description

- **Business Purpose**: Cung cấp khả năng phân loại và cô lập thông tin trên dòng hành trình sản phẩm (Timeline) dựa trên loại hình hoạt động thực tế (Ví dụ: Chỉ xem mốc kho bãi, chỉ xem mốc bán lẻ, hoặc chỉ xem mốc bảo hành). Giúp tăng cường độ tập trung thông tin khi người dùng xử lý nghiệp vụ chuyên biệt.
- **User Problem Solved**: Khi khách hàng gửi máy đi bảo hành, nhân viên sửa chữa chỉ muốn xem các sự kiện liên quan đến bảo hành (`WARRANTY_CLAIM`, `WARRANTY_RESOLVED`) của sản phẩm đó để đánh giá lỗi cũ. Họ không muốn bị rối mắt bởi hàng tá mốc vận chuyển trung chuyển trước đây.
- **Expected System Behavior**: Hệ thống nhận mảng các giá trị loại sự kiện (`eventTypes`) cần lọc, thực hiện truy vấn SELECT trong PostgreSQL với mệnh đề logic `AND event_type IN (:eventTypes)`, sau đó trả về danh sách mốc thời gian đã được thu gọn cho người dùng.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Warranty Staff | Lọc sự kiện bảo hành để nghiên cứu hồ sơ xử lý thiết bị của sản phẩm lỗi. |
| Staff Kho | Lọc sự kiện `WAREHOUSE_IN` / `WAREHOUSE_OUT` để đối chiếu biên bản vận chuyển. |
| Customer | Lọc các mốc kiểm định chất lượng để yên tâm tiêu thụ sản phẩm ăn uống. |

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

- Người dùng đang mở giao diện xem Timeline chi tiết của một sản phẩm cụ thể.
- Loại sự kiện yêu cầu lọc phải thuộc danh mục Enum được hệ thống định nghĩa.

---

# 5. Trigger

User chọn một hoặc nhiều loại sự kiện từ danh sách hộp kiểm (Checkbox list) ở thanh Sidebar bộ lọc Timeline và nhấn nút "Apply".

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại trang Timeline sản phẩm, mở bộ chọn "Event Type", tích chọn các mục "Warehouse" và "Sale", nhấn nút "Apply Filter". | Gửi yêu cầu GET tới API `/api/v1/trace/search` kèm query parameter `eventTypes=WAREHOUSE_IN,WAREHOUSE_OUT,SALE`. |
| 2 | System | Validate Enum Values | Duyệt mảng `eventTypes` nhận được, đối soát xem có phần tử nào lạ hoặc sai chính tả hay không. Nếu phát hiện sai, trả lỗi 400. |
| 3 | System | Truy vấn Database | Thực hiện câu lệnh SQL: `SELECT * FROM events WHERE product_item_id = :itemId AND event_type IN ('WAREHOUSE_IN', 'WAREHOUSE_OUT', 'SALE') ORDER BY created_at ASC`. |
| 4 | System | Ghi Audit Log | Lưu hành động lọc theo loại sự kiện của User vào log hệ thống. |
| 5 | System | Trả kết quả | Phản hồi dữ liệu JSON chứa mảng các sự kiện thỏa mãn điều kiện về cho Frontend. |
| 6 | User | Đọc thông tin | Giao diện vẽ lại dòng thời gian, chỉ thể hiện các mốc vận chuyển kho bãi và bán lẻ của sản phẩm. |

---

# 7. Alternative Flow

*Không có luồng thay thế đặc thù.*

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Loại sự kiện truyền lên sai Enum định nghĩa | Trả về HTTP 400 Bad Request: "Invalid event type filter. Supported values: PRODUCED, WAREHOUSE_IN, WAREHOUSE_OUT, SALE, REGISTERED, WARRANTY_CLAIM, WARRANTY_RESOLVED, RECALLED." |
| ERR-002 | Không tìm thấy bất kỳ mốc thời gian nào phù hợp | Hệ thống trả về mảng trống kèm mã 200 (không coi là lỗi hệ thống). |

---

# 9. Input Specification

### Request Query Parameters

```
GET /api/v1/trace/search?code=PT-MILK-SN0001&eventTypes=WAREHOUSE_IN,WAREHOUSE_OUT
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| eventTypes | String (Comma separated)| No | Phải thuộc Enum `event_type` | Danh sách các loại sự kiện cần lọc, ngăn cách bởi dấu phẩy |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "data": {
    "filterApplied": {
      "eventTypes": [
        "WAREHOUSE_IN",
        "WAREHOUSE_OUT"
      ]
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
- `events`: SELECT các sự kiện thỏa mãn bộ lọc loại hình.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `events` | SELECT | Lấy các bản ghi sự kiện có cột `event_type` nằm trong tập hợp yêu cầu |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-FTE-001 | Nếu mảng `eventTypes` gửi lên rỗng hoặc rỗng sau khi lọc sạch, hệ thống sẽ tự động coi như là xem "TẤT CẢ" (Hiển thị không lọc). |
| BR-FTE-002 | Chỉ các tài khoản vai trò Admin, Staff Kho hoặc Trung Tâm Bảo Hành mới được quyền lọc các sự kiện có tính chất nội bộ (Ví dụ: `INTERNAL_AUDIT` nếu được định nghĩa sau này). Khách hàng bình thường luôn bị chặn không cho lấy các mốc đặc thù này để tránh rò rỉ dữ liệu quy trình nghiệp vụ nội bộ. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| eventTypes | string_comma_separated | "Danh sách loại sự kiện lọc phải được ngăn cách bằng dấu phẩy hợp lệ" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/trace/search`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Trả dữ liệu mảng timeline đã lọc |
| 400 | Bad Request. Loại sự kiện gửi lên không hợp chuẩn Enum |
| 500 | Internal Server Error. Lỗi database |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình xem lịch sử sản phẩm.
- **Components**:
  - `EventTypeFilterGroup`: Thanh Sidebar chứa danh sách các Checkbox tương ứng cho từng loại mốc thời gian: "Sản xuất", "Vận chuyển", "Bán hàng", "Bảo hành", "Thu hồi".
  - Tích chọn mục nào sẽ tự động đổi màu sắc viền của mục đó để hiển thị trạng thái đang lọc tích cực.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/trace/search`
- **Responsibility**: Parse chuỗi query `eventTypes` thành mảng các chuỗi, kiểm tra tính hợp lệ của từng loại sự kiện, gọi `TraceQueryService.searchTimeline` có lọc sự kiện.

### Service
- **TraceQueryService**:
  1. Kiểm tra ID/Mã sản phẩm.
  2. Parse tham số thành mảng Postgres text array:
     ```sql
     SELECT * FROM events
     WHERE product_item_id = :itemId
       AND (:eventTypes IS NULL OR event_type = ANY(:eventTypes))
     ORDER BY created_at ASC
     ```
  3. Trả về cho Controller đóng gói.

---

# 17. Event Flow

```
[User] -> Select Event Types -> [TraceQueryService]
                                      |
                                      +---> (DB Query: WHERE event_type = ANY(eventTypes))
```

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| FILTER_TIMELINE_BY_TYPE | Staff / Customer | `userId`, `searchedCode`, `eventTypesFiltered: Array`, `timestamp` |

---

# 19. Security Consideration

- **Enum White-listing**: Bắt buộc phải hard-code mảng Enum trắng (White-list) ở Backend để đối chiếu tham số gửi lên. Tuyệt đối không nhét trực tiếp chuỗi văn bản không được lọc của user vào câu lệnh SQL nhằm chống nguy cơ SQL Injection lỗi bảo mật.

---

# 20. Acceptance Criteria

### Scenario: Lọc thành công chỉ hiển thị các sự kiện bảo hành sản phẩm
- **Given** sản phẩm điện thoại có 2 mốc kho, 1 mốc bán hàng, và 2 mốc bảo hành (`WARRANTY_CLAIM`, `WARRANTY_RESOLVED`) trong database.
- **When** nhân viên kỹ thuật tích chọn bộ lọc "Bảo hành" và bấm xem.
- **Then** hệ thống gọi API và chỉ trả về đúng 2 sự kiện bảo hành, ẩn hoàn toàn 3 sự kiện kho bãi/bán hàng đi để nhân viên dễ tập trung làm việc.

---

# 21. Developer Checklist

### Backend
- [ ] Khởi tạo mảng String Enum kiểm duyệt ở controller.
- [ ] Sửa đổi hàm query trong `TraceQueryService` để hỗ trợ mệnh đề `IN` hoặc `ANY` trong SQL.
- [ ] Đảm bảo Unit Test kiểm tra tính an toàn của đầu vào.

### Frontend
- [ ] Thiết kế bảng Checkbox đẹp mắt, có đính kèm icon minh họa kế bên chữ để tăng tính trực quan.
- [ ] Thêm nút "Select All" và "Clear All" để tăng trải nghiệm người dùng.
- [ ] Xử lý trạng thái loading khi thay đổi checkbox.
