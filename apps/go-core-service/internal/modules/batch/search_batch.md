# UC-P2-BATCH-03 - Tìm kiếm lô

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Batch |
| Use Case ID | UC-P2-BATCH-03 |
| Feature Name | Tìm kiếm lô (Batch Search API) |
| Priority | Medium |
| Git Branch | feature/batch-search |
| Producer | - |
| Consumer | Frontend |
| Dependency | UC-P1-BATCH-01 (Create Batch) |

## Description

- **Business Purpose**: Cung cấp khả năng tìm kiếm nhanh các lô sản xuất (Batch) trong hệ thống thông qua việc tìm kiếm gần đúng theo từ khóa (mã lô, tên sản phẩm mẹ, nhà sản xuất), hỗ trợ phân trang và sắp xếp kết quả linh hoạt.
- **User Problem Solved**: Khi hệ thống có hàng trăm, hàng ngàn lô sản xuất khác nhau, người dùng không thể lướt xem thủ công. Họ cần tìm nhanh một lô theo mã (ví dụ: gõ "LOT-2026") để xử lý xuất kho, kiểm định chất lượng hoặc thu hồi.
- **Expected System Behavior**: Hệ thống tiếp nhận từ khóa từ thanh tìm kiếm, xây dựng câu lệnh SQL chứa các mệnh đề `LIKE %keyword%` hoặc sử dụng index văn bản (Full-text search), thực hiện tìm kiếm trên PostgreSQL, sau đó trả về danh sách các lô hàng khớp kèm thông tin phân trang (Page, PageSize, TotalRecords).

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Admin | Tìm kiếm lô sản xuất để quản trị, theo dõi tổng thể. |
| Staff Kho | Tìm kiếm lô hàng để thực hiện các thao tác Mapping hoặc cập nhật vị trí xuất kho. |
| Dealer / Store | Tìm kiếm lô sản xuất của các sản phẩm thuộc đại lý của mình. |

---

# 3. Permission

| Role | Create | Update | Delete | View |
|---|---|---|---|---|
| Admin | No | No | No | Yes |
| Staff Kho | No | No | No | Yes |
| Dealer | No | No | No | Yes |
| Customer | No | No | No | No |

---

# 4. Preconditions

- Người dùng đã đăng nhập vào hệ thống và được cấp quyền truy cập danh sách lô hàng.
- Có dữ liệu lô hàng trong PostgreSQL (hoặc trả về danh sách rỗng nếu không có dữ liệu).

---

# 5. Trigger

User nhập từ khóa vào ô tìm kiếm tại trang Batch Management và nhấn phím Enter hoặc nút "Search".

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Truy cập trang Batch Management, gõ từ khóa tìm kiếm (ví dụ: "MILK01") và nhấn Enter. | Gửi request GET tới API `/api/v1/batches` kèm query parameter `keyword=MILK01`. |
| 2 | System | Chuẩn hóa từ khóa | Loại bỏ các khoảng trắng thừa, ký tự đặc biệt có hại (Input Sanitization) để ngăn chặn SQL Injection. |
| 3 | System | Query Database | Thực hiện truy vấn SELECT kết hợp JOIN bảng `batches` với `product_variants` để so khớp từ khóa với trường `batch_code`, `manufacturer_name`, hoặc `product_name`. Chỉ lấy các bản ghi chưa bị xóa mềm (`is_deleted = false`). |
| 4 | System | Phân trang & Sắp xếp | Áp dụng phân trang (`LIMIT` và `OFFSET`) và sắp xếp theo ngày tạo mới nhất (`created_at DESC`). |
| 5 | System | Ghi Audit Log | Ghi nhận hoạt động tìm kiếm của User vào log hệ thống. |
| 6 | System | Trả kết quả | Phản hồi JSON chứa danh sách lô hàng kèm siêu dữ liệu phân trang. |
| 7 | User | Xem danh sách | Giao diện cập nhật hiển thị danh sách các lô hàng tìm được. |

---

# 7. Alternative Flow

## AF-001 Tìm kiếm rỗng
- **Description**: Nếu người dùng không nhập gì và bấm tìm kiếm, hệ thống sẽ trả về danh sách toàn bộ các lô sản xuất hoạt động được sắp xếp theo mốc thời gian mới nhất (Mặc định).

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Từ khóa quá dài (> 100 ký tự) | Trả về HTTP 400: "Search keyword is too long. Max limit is 100 characters." |
| ERR-002 | Sai định dạng kiểu dữ liệu phân trang (Page không phải số) | Trả về lỗi: "Query parameters 'page' and 'pageSize' must be integers." |

---

# 9. Input Specification

### Request Query Parameters

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| keyword | String | No | Tối đa 100 ký tự | Từ khóa tìm theo mã lô, tên sp hoặc nhà sản xuất |
| page | Integer | No | Mặc định: `1` | Trang hiện tại cần tải |
| pageSize | Integer | No | Mặc định: `10` | Số lượng bản ghi trên mỗi trang (Max 100) |
| sortBy | String | No | Mặc định: `createdAt` | Trường sắp xếp (`createdAt`, `batchCode`) |
| sortOrder | String | No | Mặc định: `DESC` | Thứ tự sắp xếp (`ASC`, `DESC`) |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "batchId": "6f2bc881-8b21-4f10-9111-a887b2210a12",
        "batchCode": "LOT-2026-MILK01",
        "productName": "Sữa tươi tiệt trùng Vinamilk 1L",
        "manufacturingDate": "2026-07-08T00:00:00Z",
        "quantity": 5000,
        "status": "ACTIVE",
        "createdAt": "2026-07-08T18:06:00Z"
      }
    ],
    "pagination": {
      "currentPage": 1,
      "pageSize": 10,
      "totalRecords": 1,
      "totalPages": 1
    }
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `batches`: SELECT danh sách.
- `product_variants`: SELECT thông tin sản phẩm liên kết để hiển thị tên.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `batches` | SELECT | Thực hiện tìm kiếm gần đúng và đếm tổng số bản ghi phục vụ phân trang |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-SEA-001 | Kết quả tìm kiếm tuyệt đối không bao gồm các lô hàng đã bị Soft Delete (`is_deleted = true`). |
| BR-SEA-002 | Tìm kiếm không phân biệt chữ hoa, chữ thường (Case-insensitive comparison using `ILIKE` or lower conversion). |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| page | min >= 1 | "Trang hiện tại phải từ 1 trở lên" |
| pageSize | value between 1 and 100 | "Kích thước trang phải từ 1 đến 100 bản ghi" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/batches`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Trả kết quả tìm kiếm thành công (có thể trả về danh sách rỗng) |
| 400 | Bad Request. Sai định dạng query parameter |
| 401 | Unauthorized. Chưa đăng nhập |
| 500 | Internal Server Error. Gặp lỗi khi truy vấn database |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình danh sách lô hàng (`/batches`).
- **Component**:
  - `SearchBar`: Ô input có nút X để xóa nhanh từ khóa (Clear input button) và tự động focus khi tải trang.
  - `BatchTable`: Bảng hiển thị kết quả.
  - `PaginationControls`: Các nút chuyển trang (Trang đầu, trang cuối, Trang trước, Trang sau).

### UI State
- **Loading State**: Hiển thị xương dòng (Skeleton rows) mờ nhấp nháy trong bảng khi dữ liệu đang được tải từ API.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/batches`
- **Responsibility**: Parse các query parameters (`keyword`, `page`, `pageSize`, `sortBy`, `sortOrder`), gọi `BatchQueryService.searchBatches`.

### Service
- **BatchQueryService**:
  1. Kiểm tra vai trò của User để áp dụng bộ lọc dữ liệu (nếu là Dealer, chỉ lấy các lô họ được gán quyền phân phối).
  2. Xây dựng câu truy vấn SQL (Query builder) trong Postgres:
     ```sql
     SELECT b.*, pv.name as product_name
     FROM batches b
     LEFT JOIN product_variants pv ON b.variant_id = pv.id
     WHERE b.is_deleted = false
       AND (b.batch_code ILIKE :keyword OR pv.name ILIKE :keyword OR b.manufacturer_name ILIKE :keyword)
     ORDER BY b.created_at DESC
     LIMIT :limit OFFSET :offset
     ```
  3. Truy vấn song song (hoặc đếm riêng) tổng số dòng thỏa mãn điều kiện (không có limit/offset) để tính toán `totalRecords` và `totalPages`.
  4. Trả về cấu trúc DTO phân trang hoàn chỉnh.

---

# 17. Event Flow

*Không có RabbitMQ Event nào được kích hoạt vì đây là luồng đọc dữ liệu thông thường.*

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| SEARCH_BATCHES | Staff Kho / Admin | `userId`, `keyword`, `page`, `resultsCount`, `timestamp` |

---

# 19. Security Consideration

- **SQL Injection Defending**: Tuyệt đối không được ghép chuỗi string trực tiếp cho mệnh đề `WHERE` tìm kiếm. Bắt buộc dùng parameterized query (ví dụ: `:keyword` hoặc `?` thông qua ORM).

---

# 20. Acceptance Criteria

### Scenario: Tìm kiếm lô hàng bằng mã gần đúng
- **Given** trong hệ thống có 2 lô hàng có mã là `LOT-2026-MILK01` và `LOT-2027-MILK02`.
- **When** nhân viên kho gõ từ khóa "2026" vào ô tìm kiếm và bấm Enter.
- **Then** hệ thống gọi API thành công và trả về chính xác 1 kết quả là lô hàng `LOT-2026-MILK01`, không hiển thị lô còn lại.

---

# 21. Developer Checklist

### Backend
- [ ] Implement route GET `/api/v1/batches` hỗ trợ query parameters.
- [ ] Viết truy vấn `ILIKE` kết hợp join bảng sản phẩm mẹ bảo mật.
- [ ] Viết logic tính toán động metadata phân trang (`totalPages`, `totalRecords`).

### Frontend
- [ ] Thiết kế thanh Search Bar có debounce (nếu cần) hoặc nút tìm kiếm rõ ràng.
- [ ] Xử lý trường hợp "No Data Found" hiển thị minh họa bằng hình ảnh vẽ hoạt họa dễ thương.
- [ ] Tích hợp pagination đồng bộ với URL query params (để refresh trang vẫn giữ nguyên trang hiện tại).
