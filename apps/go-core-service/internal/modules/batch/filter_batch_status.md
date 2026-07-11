# UC-P2-BATCH-04 - Lọc lô theo trạng thái

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Batch |
| Use Case ID | UC-P2-BATCH-04 |
| Feature Name | Lọc lô theo trạng thái (Batch Filter API) |
| Priority | Medium |
| Git Branch | feature/batch-filter-status |
| Producer | - |
| Consumer | Frontend |
| Dependency | UC-P1-BATCH-01 (Create Batch) |

## Description

- **Business Purpose**: Cung cấp khả năng lọc nhanh chóng danh sách các lô hàng sản xuất (Batch) trong hệ thống theo trạng thái vận hành của chúng (ví dụ: đang hoạt động, đã hết hạn, đã bị thu hồi, hoặc đang bị khóa), giúp người dùng tập trung theo dõi các nhóm dữ liệu cụ thể.
- **User Problem Solved**: Khi có đợt thu hồi hàng khẩn cấp, nhân viên kho cần cô lập nhanh chóng tất cả các lô hàng có trạng thái `RECALLED` để dừng việc đóng gói và chuyển phát. Họ không muốn bị phân tâm bởi các lô hàng đang hoạt động bình thường (`ACTIVE`).
- **Expected System Behavior**: Hệ thống tiếp nhận bộ lọc trạng thái (status filter) từ request, áp dụng điều kiện `WHERE status = :status` vào câu lệnh truy vấn PostgreSQL, thực hiện phân trang, sau đó trả về danh sách các lô hàng khớp chính xác trạng thái yêu cầu.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Admin | Lọc tìm kiếm để kiểm tra hiệu suất sản xuất và quản lý thu hồi sản phẩm lỗi. |
| Staff Kho | Lọc các lô hàng `ACTIVE` để dán nhãn, mapping hoặc lọc lô `EXPIRED` để lập danh sách tiêu hủy. |
| Dealer / Store | Lọc các lô hàng thuộc cửa hàng để kiểm soát chất lượng và hạn sử dụng của hàng trưng bày. |

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

- Người dùng đã đăng nhập thành công vào hệ thống.
- Các lô sản xuất đã tồn tại trong database PostgreSQL.

---

# 5. Trigger

User nhấp chọn một trạng thái cụ thể trên thanh Tab filter (ví dụ: click Tab "Recalled") hoặc chọn trạng thái từ Dropdown bộ lọc tại trang quản lý lô.

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại màn hình Batch Management, chọn lọc theo trạng thái "Recalled" trong danh sách bộ lọc. | Gọi API `/api/v1/batches` kèm query parameter `status=RECALLED`. |
| 2 | System | Validate Enum | Kiểm tra giá trị tham số `status` xem có khớp với danh sách Enum hợp lệ của hệ thống hay không. Nếu không, ném lỗi 400. |
| 3 | System | Truy vấn Database | Thực hiện câu lệnh SQL: `SELECT * FROM batches WHERE status = 'RECALLED' AND is_deleted = false ORDER BY created_at DESC LIMIT 10 OFFSET 0`. |
| 4 | System | Ghi Audit Log | Lưu hoạt động lọc dữ liệu của User vào log hệ thống. |
| 5 | System | Trả kết quả | Phản hồi dữ liệu JSON danh sách lô hàng tương ứng về cho Frontend. |
| 6 | User | Xem thông tin | Giao diện cập nhật danh sách hiển thị, chỉ bao gồm các lô hàng bị thu hồi. |

---

# 7. Alternative Flow

## AF-001 Kết hợp lọc theo trạng thái và tìm kiếm gần đúng (Search & Filter)
- **Description**: Cho phép người dùng vừa gõ từ khóa tìm kiếm theo tên, vừa áp dụng bộ lọc trạng thái để thu hẹp kết quả tối đa (ví dụ: Tìm kiếm lô sữa có tên "Vinamilk" và đang bị `RECALLED`).

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Trạng thái gửi lên không tồn tại trong hệ thống (Enum error) | Trả về HTTP 400 Bad Request: "Invalid batch status filter value. Supported values: DRAFT, ACTIVE, EXPIRED, RECALLED, LOCKED." |
| ERR-002 | Lỗi cơ sở dữ liệu khi lọc | Trả về HTTP 500: "Internal Server Error." |

---

# 9. Input Specification

### Request Query Parameters

```
GET /api/v1/batches?status=ACTIVE&page=1&limit=10
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| status | String | No | Enum: `DRAFT`, `ACTIVE`, `EXPIRED`, `RECALLED`, `LOCKED` | Trạng thái cần lọc |
| page | Integer | No | Mặc định là `1` | Trang kết quả hiện tại |
| limit | Integer | No | Mặc định là `10`, tối đa `100` | Số lượng lô hàng mỗi trang |

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
        "variantId": "237ab291-a991-4e2a-bb33-778844ee1102",
        "quantity": 5000,
        "status": "ACTIVE",
        "expiryDate": "2027-07-08T00:00:00Z"
      }
    ],
    "pagination": {
      "currentPage": 1,
      "pageSize": 10,
      "totalRecords": 18,
      "totalPages": 2
    }
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `batches`: SELECT danh sách với bộ lọc WHERE.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `batches` | SELECT | Lấy các bản ghi lô hàng có trạng thái khớp với tham số truyền lên |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-FIL-001 | Bộ lọc trạng thái phải cho phép người dùng quay về chế độ "TẤT CẢ" (All statuses) bằng cách gửi giá trị rỗng hoặc bỏ tham số `status` khỏi query. |
| BR-FIL-002 | Chỉ các tài khoản Admin mới được lọc các lô ở trạng thái `DRAFT` (lô nháp của các cơ sở sản xuất chưa công bố chính thức). Staff Kho và Dealer không được xem lô nháp của cơ sở khác. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| status | optional, in_enum | "Giá trị bộ lọc trạng thái không hợp lệ" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/batches`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Trả về kết quả lọc thành công |
| 400 | Bad Request. Giá trị status sai định dạng Enum |
| 401 | Unauthorized. Chưa đăng nhập hệ thống |
| 500 | Internal Server Error. Gặp lỗi hệ thống |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình quản lý danh sách lô hàng (`/batches`).
- **Components**:
  - `StatusTabs`: Thanh điều hướng chứa các Tab: Tất cả, Hoạt động (`ACTIVE`), Hết hạn (`EXPIRED`), Thu hồi (`RECALLED`), Khóa (`LOCKED`), Nháp (`DRAFT`). Click vào Tab nào sẽ trigger gọi lại API với tham số `status` tương ứng.
  - `BadgeStatus`: Nhãn hiển thị màu sắc tương ứng cho từng lô hàng (ví dụ: màu đỏ cho `RECALLED`, màu xám cho `LOCKED`, màu vàng cho `EXPIRED`, màu xanh cho `ACTIVE`).

### UI State
- **Active State Styling**: Tab đang được chọn phải có màu nền đậm nổi bật hơn các Tab còn lại.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/batches`
- **Responsibility**: Đọc param `status` từ query string, gọi `BatchQueryService.searchBatches` với tham số status.

### Service
- **BatchQueryService**:
  1. Nhận thông tin vai trò người dùng hiện tại để thực thi phân quyền (nếu không phải Admin, tự động chèn thêm điều kiện `status !== 'DRAFT'`).
  2. Bổ sung điều kiện SQL động:
     ```sql
     AND b.status = :status
     ```
  3. Thực thi query lấy dữ liệu và count.
  4. Trả về kết quả phân trang.

---

# 17. Event Flow

*Không có RabbitMQ Event nào được kích hoạt vì đây là luồng đọc dữ liệu thông thường.*

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| FILTER_BATCH_STATUS | Staff Kho / Admin | `userId`, `statusFiltered`, `resultsCount`, `timestamp` |

---

# 19. Security Consideration

- **Information Disclosure Protection**: Chặn triệt để việc rò rỉ dữ liệu các lô hàng đang sửa đổi ở dạng nháp (`DRAFT`) cho các role không liên quan.

---

# 20. Acceptance Criteria

### Scenario: Lọc thành công các lô hàng bị thu hồi
- **Given** người dùng là Staff Kho đã đăng nhập thành công. Có 3 lô hàng bị thu hồi (`RECALLED`) và 10 lô hàng đang hoạt động (`ACTIVE`) trong DB.
- **When** người dùng click vào Tab "Thu hồi".
- **Then** hệ thống gọi API lọc và hiển thị chính xác danh sách gồm 3 lô hàng bị thu hồi, ẩn toàn bộ 10 lô hàng hoạt động bình thường đi.

---

# 21. Developer Checklist

### Backend
- [ ] Định nghĩa Enum chính xác cho các trạng thái của lô hàng trong code.
- [ ] Chèn logic lọc động `status` vào trong hàm query của `BatchQueryService`.
- [ ] Viết unit test kiểm tra việc ẩn các lô hàng `DRAFT` đối với tài khoản Dealer.

### Frontend
- [ ] Thiết kế components thanh Status Tabs mượt mà sử dụng `motion/react` để có hiệu ứng chuyển đổi đẹp mắt.
- [ ] Phối màu chuẩn chỉnh cho các Badge Trạng Thái (Status badge):
  - `ACTIVE`: `bg-emerald-50 text-emerald-700 border-emerald-200`
  - `RECALLED`: `bg-rose-50 text-rose-700 border-rose-200`
  - `EXPIRED`: `bg-amber-50 text-amber-700 border-amber-200`
  - `LOCKED`: `bg-slate-100 text-slate-700 border-slate-200`
- [ ] Đảm bảo chuyển đổi tab kích hoạt API gọi lại tức thì.
