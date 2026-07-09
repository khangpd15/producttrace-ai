# UC-P2-BATCH-05 - Xem sản phẩm trong lô

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Batch |
| Use Case ID | UC-P2-BATCH-05 |
| Feature Name | Xem sản phẩm trong lô (Batch Product API) |
| Priority | Medium |
| Git Branch | feature/batch-product-list |
| Producer | - |
| Consumer | Frontend |
| Dependency | UC-P1-BATCH-02 (Link Product Item to Batch) |

## Description

- **Business Purpose**: Cung cấp khả năng hiển thị chi tiết danh sách tất cả các sản phẩm đơn lẻ (Product Items) đã được đóng gói và gán nhãn thuộc về một lô sản xuất (Batch) cụ thể, hỗ trợ đắc lực cho công tác kiểm kho và theo dõi phân phối.
- **User Problem Solved**: Khi có 5,000 hộp sữa trong một lô, nhân viên không biết có những số Serial nào đã được in, những sản phẩm nào đã bán đi, và những sản phẩm nào đang nằm tại Đại lý nào trong chuỗi cung ứng.
- **Expected System Behavior**: Hệ thống tiếp nhận ID của lô hàng, truy vấn toàn bộ các dòng dữ liệu trong bảng `product_items` có trường `batch_id` khớp, kết xuất thông tin chi tiết của từng sản phẩm (bao gồm mã QR, serial, trạng thái, vị trí hiện tại) và trả về Frontend dưới dạng bảng dữ liệu có phân trang.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Staff Kho | Kiểm đếm thực tế danh sách các số Serial trong lô để đối soát trước khi vận chuyển. |
| Admin / Supervisor | Theo dõi tỷ lệ bán hàng và phân phối thực tế của từng sản phẩm trong lô hàng. |
| Dealer / Store | Kiểm tra trạng thái của các sản phẩm riêng lẻ thuộc lô hàng mà họ đã nhận. |

---

# 3. Permission

| Role | Create | Update | Delete | View |
|---|---|---|---|---|
| Admin | No | No | No | Yes |
| Staff Kho | No | No | No | Yes |
| Dealer | No | No | No | Yes* |
| Customer | No | No | No | No |

(*) Dealer chỉ được xem các sản phẩm trong lô đang nằm tại kho hoặc cửa hàng thuộc quyền sở hữu/quản lý của họ.

---

# 4. Preconditions

- Lô sản xuất liên quan phải tồn tại trong PostgreSQL.
- Người dùng đã đăng nhập và được cấp quyền xem danh mục sản phẩm đơn lẻ (`VIEW_BATCH_PRODUCTS`).

---

# 5. Trigger

User truy cập màn hình chi tiết lô sản xuất (Batch Detail) và click vào Tab "Danh sách sản phẩm đơn lẻ" hoặc click nút "View Items".

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại trang chi tiết Batch, chọn Tab "Products in Batch". | Gửi yêu cầu GET tới API `/api/v1/batches/:batchId/products` kèm các tham số phân trang. |
| 2 | System | Xác thực Token | Kiểm tra JWT token, vai trò người dùng và phạm vi dữ liệu được phép truy cập (Data scoping). |
| 3 | System | Truy vấn Database | Thực hiện JOIN bảng `product_items` với bảng `locations` để lấy tên địa điểm lưu kho hiện tại của từng hộp sữa. |
| 4 | System | Phân trang & Lọc | Áp dụng phân trang bắt buộc (mặc định 20 dòng) để bảo vệ hiệu năng hệ thống. Sắp xếp theo thứ tự `serial_number ASC` hoặc `item_code ASC`. |
| 5 | System | Ghi Audit Log | Ghi nhận hành động xem danh sách sản phẩm của lô vào bảng `audit_logs`. |
| 6 | System | Trả kết quả | Phản hồi dữ liệu JSON danh sách sản phẩm đơn lẻ về cho Frontend. |
| 7 | User | Xem dữ liệu | Giao diện hiển thị bảng danh sách các hộp sữa có kèm nhãn trạng thái và vị trí tương ứng. |

---

# 7. Alternative Flow

## AF-001 Tìm kiếm nhanh sản phẩm trong lô (Search items in batch)
- **Description**: Cho phép nhân viên gõ trực tiếp số Serial vào ô tìm kiếm nhỏ trong Tab này để lọc nhanh xem hộp sữa có số Serial đó có nằm trong lô sản xuất này không.

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Lô sản xuất không tồn tại hoặc đã bị xóa | Trả về HTTP 404: "Batch not found." |
| ERR-002 | Lô sản xuất hoàn toàn trống rỗng (chưa có sản phẩm nào được gán) | Hệ thống trả về mảng `items` rỗng kèm mã 200 (không coi là lỗi hệ thống). |

---

# 9. Input Specification

### Request Query Parameters

```
GET /api/v1/batches/6f2bc881-8b21-4f10-9111-a887b2210a12/products?page=1&limit=20&status=AVAILABLE
```

### Table Specification

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| page | Integer | No | Mặc định: `1` | Số thứ tự trang cần lấy |
| limit | Integer | No | Mặc định: `20`, tối đa `100` | Số lượng sản phẩm hiển thị mỗi trang |
| status | String | No | Enum: `AVAILABLE`, `IN_TRANSIT`, `SOLD`, `RECALLED` | Lọc theo trạng thái hộp sữa |
| keyword | String | No | Tối đa 50 ký tự | Tìm nhanh theo số Serial hoặc mã sản phẩm |

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "itemId": "c20befe6-e4b9-4f37-b514-e66233ef04a1",
        "itemCode": "PT-MILK-SN0001",
        "serialNumber": "SN-2026-0001",
        "status": "AVAILABLE",
        "currentLocation": {
          "locationId": "1a1fa1a0-1200-4b2e-a551-fb112aaee088",
          "name": "Kho tổng Vinamilk Bình Dương",
          "type": "WAREHOUSE"
        },
        "createdAt": "2026-07-08T18:06:00Z"
      }
    ],
    "pagination": {
      "currentPage": 1,
      "pageSize": 20,
      "totalRecords": 4500,
      "totalPages": 225
    }
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `product_items`: SELECT danh sách sản phẩm.
- `locations`: SELECT thông tin tên kho/cửa hàng lưu kho hiện tại.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `product_items`| SELECT | Lấy các hộp sản phẩm có `batch_id = :batchId` |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-BPR-001 | Danh sách sản phẩm bắt buộc phải có phân trang (Pagination). Nghiêm cấm việc select không giới hạn (Unlimited select) vì một lô hàng có thể chứa hàng chục nghìn sản phẩm gây treo ram máy chủ và sập trình duyệt của người dùng. |
| BR-BPR-002 | Nếu tài khoản là Đại lý (Dealer), hệ thống bắt buộc phải chèn thêm điều kiện SQL lọc: `AND currentLocationId IN (danh sách kho thuộc quản lý của Dealer)`. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| page | min 1 | "Số trang phải bắt đầu từ 1" |
| limit | min 1, max 100 | "Giới hạn số bản ghi mỗi trang từ 1 đến 100" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/batches/:batchId/products`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Trả dữ liệu danh sách sản phẩm thành công |
| 401 | Unauthorized. Chưa đăng nhập |
| 403 | Forbidden. Không có quyền xem danh sách sản phẩm thuộc lô hàng này |
| 404 | Not Found. Không tìm thấy lô sản xuất tương ứng |
| 500 | Internal Server Error. Lỗi database |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình chi tiết lô hàng (`/batches/:id`).
- **Component**: `ProductItemTable` - Bảng dữ liệu gọn gàng, hiển thị mã sản phẩm, Serial Number, Badge trạng thái có màu và cột hiển thị vị trí hiện tại (ví dụ: "Cửa hàng Quận 1"). Có thanh bộ lọc nhanh theo Trạng thái (Dropdown) và ô Search Serial.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/batches/:batchId/products`
- **Responsibility**: Nhận `batchId` và các query parameters, gọi `BatchQueryService.getProductsInBatch`.

### Service
- **BatchQueryService**:
  1. Kiểm tra sự tồn tại của Batch trong Postgres.
  2. Phân tích quyền hạn của Actor để bổ sung filter vị trí (nếu là Dealer).
  3. Xây dựng câu lệnh SQL có phân trang:
     ```sql
     SELECT pi.id as "itemId", pi.item_code as "itemCode", pi.serial_number as "serialNumber", pi.status,
            l.id as "locationId", l.name as "locationName", l.type as "locationType"
     FROM product_items pi
     LEFT JOIN locations l ON pi.current_location_id = l.id
     WHERE pi.batch_id = :batchId
       AND pi.is_deleted = false
     ORDER BY pi.serial_number ASC
     LIMIT :limit OFFSET :offset
     ```
  4. Thực hiện đếm tổng số dòng để hoàn thiện object phân trang.
  5. Trả về cho Controller.

---

# 17. Event Flow

*Không có luồng sự kiện đẩy lên hàng đợi RabbitMQ vì đây là API truy vấn đọc dữ liệu thông thường.*

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| VIEW_BATCH_PRODUCTS | Staff Kho / Admin / Dealer | `userId`, `batchId`, `page`, `pageSize`, `timestamp` |

---

# 19. Security Consideration

- **SQL Offset Attack Prevention**: Giới hạn tối đa kích thước offset có thể truy cập để tránh việc kẻ tấn công cố tình duyệt trang quá sâu (ví dụ: page = 100000) gây chậm cơ sở dữ liệu.

---

# 20. Acceptance Criteria

### Scenario: Xem danh sách sản phẩm của lô sữa tươi thành công
- **Given** lô sản xuất `LOT-001` có 100 hộp sữa đơn lẻ đã được liên kết thành công.
- **When** nhân viên kho mở màn hình chi tiết lô và chuyển sang tab danh sách sản phẩm.
- **Then** hệ thống hiển thị bảng dữ liệu chứa danh sách các hộp sữa có số Serial từ `SN-0001` đến `SN-0020` ở trang 1, cột vị trí hiện tại hiển thị đúng là "Kho tổng Bình Dương".

---

# 21. Developer Checklist

### Backend
- [ ] Thiết lập API GET `/api/v1/batches/:batchId/products`.
- [ ] Tích hợp join bảng `locations` lấy tên địa điểm lưu trữ.
- [ ] Viết logic phân trang chặt chẽ, tối ưu hóa tốc độ query bằng cách đánh index trên trường `batch_id` của bảng `product_items`.

### Frontend
- [ ] Xây dựng bảng hiển thị danh sách sản phẩm đơn lẻ responsive.
- [ ] Thêm các Badge hiển thị màu sắc tương ứng cho trạng thái của từng hộp sữa.
- [ ] Thiết kế thanh tìm kiếm nhanh số Serial trong bảng không bị lag (nếu có thể thì kết hợp debounce 300ms).
