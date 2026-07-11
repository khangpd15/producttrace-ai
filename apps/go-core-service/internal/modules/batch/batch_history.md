# UC-P2-BATCH-06 - Xem lịch sử thay đổi lô

# 1. Overview

## Basic Information

| Field | Value |
|---|---|
| Domain | Batch |
| Use Case ID | UC-P2-BATCH-06 |
| Feature Name | Xem lịch sử thay đổi lô (Batch History API) |
| Priority | Medium |
| Git Branch | feature/batch-history |
| Producer | batch.history_viewed |
| Consumer | Dashboard |
| Dependency | UC-P2-BATCH-01 (Update Batch) |

## Description

- **Business Purpose**: Cung cấp khả năng truy vết và kiểm soát toàn bộ nhật ký thay đổi thông tin (Audit trail) của một lô sản xuất cụ thể trong suốt vòng đời của nó. Giúp doanh nghiệp bảo đảm tính chịu trách nhiệm và phát hiện sai sót dữ liệu.
- **User Problem Solved**: Khi thông tin của một lô hàng bị thay đổi (ví dụ: ngày hết hạn bị đẩy lùi lại, hoặc trạng thái đột ngột chuyển sang `ACTIVE` dù đang bị `RECALLED`), người quản lý cần biết ai đã thực hiện thay đổi đó, vào thời gian nào, và giá trị cũ là gì để đối soát.
- **Expected System Behavior**: Hệ thống tiếp nhận yêu cầu xem lịch sử, truy vấn bảng logs/audit_logs theo `batch_id`, gom tất cả các hành động (CREATE, UPDATE, STATUS_CHANGE) xếp theo thứ tự thời gian giảm dần, và trả về cho Frontend hiển thị dưới dạng bảng đối chiếu.

---

# 2. Actors

| Actor | Responsibility |
|---|---|
| Admin | Xem lịch sử thay đổi để kiểm duyệt, quản lý chất lượng và phát hiện gian lận hoặc sai sót của nhân viên. |
| Staff Kho | Tra cứu lịch sử sửa đổi để biết lý do cập nhật thông tin của các lô hàng họ quản lý. |
| Dashboard Service | Consumer nhận sự kiện để tổng hợp số lượng lượt xem/truy xuất dữ liệu của lô lên màn hình quản trị chung. |

---

# 3. Permission

| Role | Create | Update | Delete | View |
|---|---|---|---|---|
| Admin | No | No | No | Yes |
| Staff Kho | No | No | No | Yes |
| Dealer | No | No | No | No |
| Customer | No | No | No | No |

---

# 4. Preconditions

- Người dùng đã đăng nhập hệ thống và được cấp quyền `VIEW_BATCH_HISTORY`.
- Lô sản xuất liên quan phải tồn tại trong PostgreSQL.

---

# 5. Trigger

User nhấn nút "View History" hoặc biểu tượng chiếc đồng hồ ngược trên màn hình chi tiết lô hàng (Batch Detail).

---

# 6. Main Workflow

| Step | Actor | Action | System Behavior |
|---|---|---|---|
| 1 | User | Tại trang chi tiết lô, click vào nút "Change History". | Điều hướng và gửi request GET đến API `/api/v1/batches/:batchId/history`. |
| 2 | System | Xác thực Token | Kiểm tra JWT token và quyền hạn xem lịch sử hệ thống của tài khoản. |
| 3 | System | Truy vấn Lịch sử | Truy vấn trong bảng `audit_logs` tất cả các bản ghi có cột `entity_type = 'BATCH'` và `entity_id = :batchId`. |
| 4 | System | Sắp xếp | Sắp xếp kết quả theo thời gian `created_at DESC` để hiển thị các thay đổi mới nhất ở trên cùng. |
| 5 | System | Ghi sự kiện | Publish message `batch.history_viewed` lên RabbitMQ exchange `batch.events`. |
| 6 | System | Trả dữ liệu | Trả về JSON chứa mảng các bản ghi lịch sử thay đổi kèm theo tên người thực hiện thay đổi. |
| 7 | User | Xem đối chiếu | Giao diện hiển thị bảng đối chiếu rõ ràng: "Trường thay đổi", "Giá trị cũ", "Giá trị mới", "Người thực hiện", "Thời gian". |

---

# 7. Alternative Flow

*Không có luồng thay thế. Luồng xem lịch sử là cố định từ bảng Audit Logs.*

---

# 8. Exception Handling

| Error Code | Scenario | System Response |
|---|---|---|
| ERR-001 | Lô sản xuất không tồn tại trong hệ thống | Trả về HTTP 404: "Batch not found." |
| ERR-002 | Lô sản xuất chưa từng có lịch sử thay đổi (Chỉ có sự kiện tạo ban đầu) | Trả về danh sách chứa 1 dòng lịch sử loại `CREATE` với mã HTTP 200. |

---

# 9. Input Specification

### Request Parameter (URL Path)

- **Parameter**: `batchId` (UUID) - Mã ID lô hàng.

### Request Query Parameters
- `page` (Integer) - Mặc định `1`.
- `limit` (Integer) - Mặc định `15`.

---

# 10. Output Specification

### Response DTO (JSON)

```json
{
  "success": true,
  "data": {
    "batchId": "6f2bc881-8b21-4f10-9111-a887b2210a12",
    "batchCode": "LOT-2026-MILK01",
    "history": [
      {
        "logId": "aa9cc812-7bb1-4110-8aa2-9f881b2a9912",
        "action": "UPDATE",
        "changedFields": {
          "expiryDate": {
            "old": "2027-01-01T00:00:00Z",
            "new": "2027-07-08T00:00:00Z"
          },
          "status": {
            "old": "DRAFT",
            "new": "ACTIVE"
          }
        },
        "performedBy": {
          "userId": "a20befe6-e4b9-4f37-b514-e66233ef04a1",
          "fullName": "Trần Văn B",
          "role": "ADMIN"
        },
        "ipAddress": "192.168.1.15",
        "createdAt": "2026-07-08T18:06:00Z"
      }
    ]
  }
}
```

---

# 11. Database Impact

### Tables Affected
- `audit_logs`: SELECT danh sách lịch sử sửa đổi.
- `users`: SELECT thông tin tên và chức vụ của người sửa.

### CRUD Operation

| Table | Operation | Description |
|---|---|---|
| `audit_logs`| SELECT | Truy vấn các dòng audit log gán với thực thể lô hàng |

---

# 12. Business Rules

| Rule ID | Rule |
|---|---|
| BR-HIS-001 | Bảng đối sánh thay đổi phải thể hiện rõ dạng so sánh cũ-mới (Diff-viewer structure) để người dùng dễ đọc nhất. |
| BR-HIS-002 | Lịch sử thay đổi là dữ liệu chỉ đọc (Immutable records). Không một ai, kể cả Admin được quyền chỉnh sửa hay xóa các dòng nhật ký audit_logs này. |

---

# 13. Validation Rules

| Field | Rule | Message |
|---|---|---|
| batchId | required, UUID format | "ID lô sản xuất không đúng cấu trúc UUID" |

---

# 14. API Design

### Endpoint

- **Method**: GET
- **Path**: `/api/v1/batches/:batchId/history`

### HTTP Status

| Status | Meaning |
|---|---|
| 200 | OK. Trả về lịch sử thay đổi thành công |
| 401 | Unauthorized. Chưa đăng nhập hệ thống |
| 403 | Forbidden. Không có quyền xem audit log |
| 404 | Not Found. Không tìm thấy lô hàng |

---

# 15. Frontend Requirement

### Page & Component
- **Page**: Màn hình xem chi tiết lịch sử lô hàng.
- **Component**: `HistoryDiffViewer` - Bảng dữ liệu có giao diện so sánh:
  - Cột Giá trị cũ được tô đỏ nhạt (màu đỏ gạch xóa - strike through).
  - Cột Giá trị mới được tô xanh nhạt (màu xanh lá tươi sáng).
  - Chứa cột ghi thông tin IP và hệ điều hành của người sửa để tăng tính bảo mật.

---

# 16. Backend Implementation Guide

### Controller
- **Route**: GET `/api/v1/batches/:batchId/history`
- **Responsibility**: Validate `batchId` từ URL path, gọi `AuditLogQueryService.getBatchHistory`.

### Service
- **AuditLogQueryService**:
  1. Thực hiện query bảng `audit_logs` join bảng `users`.
  2. Map cấu trúc dữ liệu JSON thô lưu trong cột `old_values` và `new_values` thành cấu trúc cặp key-value đối chiếu cho Frontend.
  3. Publish event `batch.history_viewed` lên RabbitMQ để Dashboard nhận biết.
  4. Trả kết quả về cho Controller.

### Event Payload (RabbitMQ)

- **Exchange**: `dashboard.events`
- **Routing Key**: `batch.history_viewed`
- **Payload**:
```json
{
  "eventId": "gg9cc812-7bb1-4110-8aa2-9f881b2a99aa",
  "eventType": "batch.history_viewed",
  "timestamp": "2026-07-08T18:06:00Z",
  "data": {
    "batchId": "6f2bc881-8b21-4f10-9111-a887b2210a12",
    "viewedBy": "a20befe6-e4b9-4f37-b514-e66233ef04a1"
  }
}
```

---

# 17. Event Flow

```
[Admin / Manager] -> Request History -> [AuditLogQueryService]
                                              |
                                              +---> (DB Read audit_logs)
                                              |
                                              +---> Publish history_viewed -> [RabbitMQ Broker]
                                                                                   |
                                                                                   +---> Consumer: [Dashboard Service]
```

---

# 18. Audit Log Requirement

| Action | Actor | Data Logged |
|---|---|---|
| VIEW_BATCH_CHANGE_HISTORY| Admin / Staff | `userId`, `batchId`, `ipAddress`, `timestamp` |

---

# 19. Security Consideration

- **Masking Sensitive Data**: Nếu trong lịch sử thay đổi có chứa thông tin nhạy cảm của hệ thống (ví dụ: token bảo mật hay giá vốn hàng bán), hệ thống phải thực hiện che mờ (masking) bằng dấu hoa thị `***` trước khi trả ra Frontend.

---

# 20. Acceptance Criteria

### Scenario: Hiển thị bảng so sánh giá trị cũ và mới khi sửa hạn dùng
- **Given** Admin vừa sửa hạn dùng của lô `LOT-S1` từ ngày `2026-12-31` thành ngày `2027-06-30`.
- **When** người quản lý kho bấm vào trang Change History của lô này.
- **Then** hệ thống hiển thị một dòng thay đổi của trường `expiryDate`, giá trị cũ hiển thị gạch ngang `2026-12-31`, giá trị mới hiển thị màu xanh là `2027-06-30`, do Admin thực hiện.

---

# 21. Developer Checklist

### Backend
- [ ] Thiết lập API GET `/api/v1/batches/:batchId/history`.
- [ ] Đảm bảo dữ liệu ghi vào `audit_logs` khi sửa lô ở UC-P2-BATCH-01 có cấu trúc JSON hợp chuẩn dễ map.
- [ ] Viết bộ parser JSON sang cấu trúc diff-view ở service.

### Frontend
- [ ] Thiết kế bảng so sánh màu sắc đỏ/xanh lá chuẩn mực.
- [ ] Thêm thông tin người chỉnh sửa và ngày giờ chi tiết.
- [ ] Xử lý phân trang cho danh sách lịch sử quá dài.
