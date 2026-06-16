# TÀI LIỆU CẤU TRÚC DỮ LIỆU (DATA DICTIONARY)

Dự án: ProductTrace AI

Phiên bản: 1.1 (Cập nhật bổ sung bảng product_items)

Vai trò tài liệu: Mô tả cấu trúc và ý nghĩa nghiệp vụ của cơ sở dữ liệu (PostgreSQL), đối chiếu với tài liệu BRD.

## 1. Database Overview

Cơ sở dữ liệu của ProductTrace AI được thiết kế để lưu trữ, quản lý và truy vết toàn bộ vòng đời của sản phẩm từ nhà máy đến tay người tiêu dùng cuối.

Hệ thống được chia thành 5 domain (phân hệ) nghiệp vụ chính:

1. User & Access Management: Quản lý người dùng, phân quyền (RBAC) cho nội bộ và khách hàng.
2. Product Catalog: Quản lý danh mục, thông tin sản phẩm chung, biến thể (Variant) và thuộc tính động (Attributes).
3. Inventory & Entities: Quản lý đối tượng vật lý bao gồm Lô hàng (Batch), Sản phẩm cụ thể (Product Item - định danh bằng Serial/QR), và Điểm vật lý (Shops/Warehouse).
4. Traceability (Truy vết): Ghi nhận lịch sử (Timeline) bất biến của sản phẩm qua các khâu sản xuất, vận chuyển, bán hàng.
5. Post-sale (Hậu mãi): Quản lý quyền sở hữu sản phẩm (Ownership) và vòng đời bảo hành (Warranty).

## 2. Table Description

### Phân hệ: User & Access Management

#### Table: users

* Business Purpose: Lưu trữ thông tin của tất cả các tác nhân tương tác với hệ thống (Khách hàng, Nhân viên, Đại lý, Admin).
* Used In Features: Đăng ký/Đăng nhập (Auth), Phân quyền (RBAC), Đăng ký sở hữu, Kích hoạt bảo hành.
* Relationships: Một User có thể sở hữu nhiều product_items, quản lý nhiều shops, tạo nhiều products/batches, và thực hiện nhiều events.



|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| id | uuid | YES | Định danh duy nhất của người dùng. | 550e8400-e29b-... | Khóa chính |
| email | varchar | YES | Tên đăng nhập và kênh liên lạc chính. | khachhang@gmail.com | Unique. Đã được đánh index (lower case) |
| phone | varchar | NO | Số điện thoại liên lạc/nhận OTP. | 0901234567 | Unique |
| full_name | text | YES | Tên hiển thị trên hệ thống/phiếu bảo hành. | Nguyễn Văn A |  |
| password_hash | varchar | NO | Mật khẩu mã hóa. Có thể NULL nếu dùng Social Login. | ... | Kỹ thuật |
| role | varchar | NO | Phân định quyền hạn trong hệ thống (RBAC). | CUSTOMER | Mặc định: CUSTOMER |
| status | varchar | NO | Trạng thái tài khoản để kiểm soát truy cập. | ACTIVE | Mặc định: ACTIVE |
| avatar_url | text | NO | Ảnh đại diện của người dùng. | https://.../avatar.jpg |  |
| is_deleted | boolean | NO | Đánh dấu xóa mềm, giữ lại data phục vụ audit/truy vết. | false | Soft delete |

### Phân hệ: Product Catalog

#### Table: product_categories

* Business Purpose: Cây phân cấp danh mục sản phẩm (Ví dụ: Điện tử -> Điện thoại -> Smartphone).
* Used In Features: Quản lý sản phẩm, Bộ lọc tìm kiếm Vector Search.
* Relationships: Một danh mục có thể có một danh mục cha (parent_id). Một danh mục chứa nhiều products và attributes.

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| parent_id | uuid | NO | Xác định danh mục cha để tạo cấu trúc cây. | uuid... | NULL nếu là Root Category |
| code | varchar | NO | Mã danh mục dùng nội bộ hoặc tích hợp ERP. | ELEC_PHONE | Unique |
| name | varchar | YES | Tên danh mục hiển thị cho khách hàng. | Điện thoại thông minh |  |
| slug | varchar | NO | URL thân thiện dùng cho SEO / PWA. | dien-thoai-thong-minh | Unique |

#### Table: products

* Business Purpose: Thông tin chung của một dòng sản phẩm (Master Data).
* Used In Features: Admin quản lý Catalog, Hiển thị thông tin chung khi quét QR.
* Relationships: Thuộc 1 product_categories. Có nhiều product_variants.

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| category_id | uuid | YES | Phân loại sản phẩm thuộc nhóm nào. | uuid... |  |
| name | varchar | YES | Tên dòng sản phẩm. | iPhone 16 Pro Max |  |
| tags | jsonb | NO | Các từ khóa hỗ trợ tìm kiếm và AI Vector Search. | ["apple", "flagship"] | Dùng GIN Index |
| metadata_json | jsonb | NO | Thông tin mở rộng không cố định cấu trúc (VD: Nguồn gốc thương hiệu). | {"brand": "Apple"} |  |
| status | varchar | NO | Trạng thái kinh doanh của dòng sản phẩm này. | ACTIVE |  |

#### Table: product_variants

* Business Purpose: Phiên bản cụ thể của sản phẩm để bán (SKU). Một Product có thể có nhiều Variant (Khác màu, khác dung lượng).
* Used In Features: Bán hàng, Định giá, Hiển thị chi tiết SP khi quét QR.

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| product_id | uuid | YES | Variant này thuộc dòng sản phẩm nào. | uuid... |  |
| sku | varchar | YES | Mã lưu kho quản lý nội bộ. | IP16PM-256-TITAN | Unique |
| barcode | varchar | NO | Mã vạch quốc tế (EAN/UPC) dùng để quét tại quầy. | 8931234567890 | Unique |
| price | decimal | NO | Giá bán lẻ niêm yết (Dùng để filter khi tìm kiếm). | 30000000 | Cần ≥ 0 |
| images_json | jsonb | NO | Danh sách ảnh chụp chi tiết của phiên bản này. | ["img1.jpg", "img2.jpg"] |  |

#### Table: attributes & attribute_values

* Business Purpose: Lưu trữ cấu hình thông số kỹ thuật động (Dynamic EAV model). Thay vì tạo thêm cột trong DB cho mỗi loại SP, hệ thống lưu theo dạng Key-Value (Ví dụ: RAM: 8GB).
* Used In Features: Hiển thị thông số kỹ thuật (FR-04).

### Phân hệ: Inventory & Entities

#### Table: batches

* Business Purpose: Lô sản xuất. Quản lý hạn sử dụng và phục vụ tính năng thu hồi (Recall) hàng loạt.
* Used In Features: Quản lý lô (FR-02), Cảnh báo sắp hết hạn (RPT-04), Báo cáo thu hồi (RPT-05).

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| variant_id | uuid | YES | Lô này sản xuất ra loại sản phẩm (SKU) nào. | uuid... |  |
| batch_code | varchar | YES | Mã lô do nhà máy cung cấp. | LOT-2026-06A | Unique |
| manufacture_date | timestamp | NO | Ngày sản xuất của cả lô. | 2026-06-01 |  |
| expiry_date | timestamp | NO | Hạn sử dụng (Quan trọng cho ngành FMCG/Dược). | 2028-06-01 | Phải ≥ manufacture_date |
| quantity | integer | NO | Tổng số lượng sản phẩm được tạo ra trong lô này. | 5000 | Mặc định 0 |
| status | varchar | NO | Tình trạng của lô (Cho phép lưu hành hay Thu hồi). | ACTIVE |  |

#### Table: shops

* Business Purpose: Điểm kiểm soát vật lý. Có thể là nhà máy, kho trung tâm, đại lý bán lẻ, hoặc trung tâm bảo hành.
* Used In Features: Ghi nhận vị trí (Timeline event), Geo Search tìm điểm gần nhất (FR-12).

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| type | varchar | NO | Loại hình điểm vật lý (Kho, Cửa hàng, TT Bảo hành). | WARRANTY_CENTER |  |
| address | text | NO | Địa chỉ chi tiết. | 123 Lê Lợi, Q1 |  |
| geo_location | geography | NO | Tọa độ không gian (Point) để phục vụ định vị GPS. | POINT(106.7 10.7) | PostGIS (Giúp query siêu nhanh) |
| opening_hours_json | jsonb | NO | Lịch làm việc phục vụ KH tra cứu điểm bảo hành. | {"Mon-Fri":"8h-17h"} |  |

#### Table: product_items

* Business Purpose: Quản lý từng sản phẩm vật lý DUY NHẤT có QR. Mỗi item có item_code (PTA-YYMM-XXXXXXXX) và verification_token (HMAC-SHA256) tạo thành QR payload xác thực chống hàng giả. Đây là thực thể được theo dõi xuyên suốt vòng đời.
* Used In Features: Quản lý QR Code (BR-01, BR-06), Truy vết cá thể, Đăng ký bảo hành.

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| id | uuid | YES | Định danh duy nhất trong database. | uuid... | Khóa chính |
| variant_id | uuid | YES | Phiên bản của sản phẩm này. | uuid... | FK → product_variants |
| batch_id | uuid | NO | Sản phẩm này thuộc lô nào. | uuid... | FK → batches |
| current_shop_point_id | uuid | NO | Sản phẩm hiện đang nằm ở kho/cửa hàng nào. | uuid... | FK → shops |
| item_code | varchar(20) | YES | Mã định danh sản phẩm. | PTA-2606-12345678 | Unique. Format: PTA-YYMM-XXXXXXXX |
| verification_token | varchar(32) | YES | Token bảo mật HMAC-SHA256 truncated, dùng để validate khi scan QR. | a1b2c3d4e5... | Đảm bảo tính chống hàng giả |
| serial_number | varchar | NO | Số Serial do nhà máy in trên vỏ hộp. | SN-A1B2C3 | Unique |
| status | varchar | NO | Trạng thái hiện tại trong vòng đời. | IN_STOCK | IN_STOCK, IN_TRANSIT, AT_DEALER, SOLD, REGISTERED,... |
| produced_at | timestamp | NO | Thời điểm sản phẩm được sản xuất. | 2026-06-01 10:00:00 |  |
| packed_at | timestamp | NO | Thời điểm sản phẩm được đóng gói. | 2026-06-02 10:00:00 |  |
| sold_at | timestamp | NO | Thời điểm bán ra cho khách hàng. | 2026-06-15 10:00:00 |  |
| registered_at | timestamp | NO | Thời điểm khách đăng ký sở hữu. | 2026-06-16 10:00:00 |  |
| last_scanned_at | timestamp | NO | Thời điểm QR được quét gần đây nhất. | 2026-06-17 10:00:00 | Phục vụ tracking và detect bất thường |
| metadata_json | jsonb | NO | Dữ liệu mở rộng, tùy biến của sản phẩm. | {"qc_passed": true} |  |
| created_at | timestamp | NO | Thời điểm tạo record. | 2026-06-01 10:00:00 |  |
| updated_at | timestamp | NO | Thời điểm cập nhật record. | 2026-06-15 10:00:00 |  |
| is_deleted | boolean | NO | Đánh dấu xóa mềm. | false |  |

### Phân hệ: Traceability (Truy vết)

#### Table: events

* Business Purpose: Nhật ký vòng đời sản phẩm. Bất biến (Immutable) theo chuẩn BR-02. Trả lời câu hỏi: Ai đã làm gì, với cái gì, ở đâu, khi nào?
* Used In Features: Traceability Timeline (FR-05, FR-06).

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| product_item_id | uuid | NO | Áp dụng cho một sản phẩm cụ thể. | uuid... |  |
| batch_id | uuid | NO | Hoặc áp dụng cho cả lô (VD: Nhập kho cả lô). | uuid... | Check: Phải có Item hoặc Batch |
| actor_id | uuid | NO | Người/Tài khoản thực hiện thao tác. | uuid... | Phục vụ Audit |
| shop_point_id | uuid | NO | Nơi xảy ra sự kiện. | uuid... |  |
| event_type | varchar | YES | Phân loại hành động truy vết. | WAREHOUSE_IN |  |
| geo_location | geography | NO | Tọa độ GPS lúc diễn ra hành động scan/nhập liệu. | POINT(...) | Chống gian lận vị trí |

#### Table: attachments

* Business Purpose: Lưu trữ file bằng chứng đi kèm với các sự kiện.
* Used In Features: Upload ảnh/video khi yêu cầu bảo hành (FR-09), Chụp ảnh hóa đơn bán hàng.

### Phân hệ: Post-sale (Hậu mãi)

#### Table: ownerships

* Business Purpose: Ghi nhận và bảo vệ quyền sở hữu pháp lý của một người với một sản phẩm.
* Used In Features: Kích hoạt sở hữu (BR-03), Chuyển nhượng (BR-15), Lịch sử sở hữu (RPT-06).

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| product_item_id | uuid | YES | Sản phẩm được sở hữu. | uuid... |  |
| owner_id | uuid | YES | Khách hàng là chủ sở hữu. | uuid... |  |
| status | varchar | NO | Trạng thái hiệu lực của chứng nhận sở hữu. | ACTIVE | Để thỏa BR-03, chỉ có 1 ACTIVE/Item |
| purchase_date | timestamp | NO | Ngày khách hàng mua hàng (Cơ sở tính bảo hành). | 2026-06-15 |  |

#### Table: warranties

* Business Purpose: Sổ bảo hành điện tử của thiết bị. Phụ thuộc vào ownerships.
* Used In Features: Quản lý bảo hành, Tạo claim lỗi (FR-08, FR-09, FR-10).

|  |  |  |  |  |  |
| --- | --- | --- | --- | --- | --- |
| Column | Data Type | Required | Business Meaning | Example Value | Notes |
| product_item_id | uuid | YES | Thiết bị được bảo hành. | uuid... |  |
| warranty_code | varchar | NO | Mã số bảo hành để tra cứu nhanh. | WAR-123456 | Unique |
| status | varchar | NO | Tình trạng gói bảo hành. | ACTIVE |  |
| start_date | timestamp | NO | Ngày bắt đầu tính hiệu lực bảo hành. | 2026-06-15 |  |
| end_date | timestamp | NO | Ngày hết hạn. | 2027-06-15 |  |

## 3. Enumerations / Status Definitions (Luật Trạng thái)

Hệ thống sử dụng các CHECK Constraints để bảo vệ tính toàn vẹn của dữ liệu trạng thái.

### 3.1. User Roles (users.role)

|  |  |
| --- | --- |
| Value | Meaning |
| ADMIN | Quản trị viên hệ thống, toàn quyền xem báo cáo, quản trị người dùng. |
| STAFF | Kỹ thuật viên bảo hành, thủ kho (thực hiện nghiệp vụ vận hành nội bộ). |
| DEALER | Đối tác/Cửa hàng ủy quyền (Quét mã nhập/xuất kho, bán hàng). |
| CUSTOMER | Khách hàng cuối (Quét QR, đăng ký sở hữu, yêu cầu bảo hành). |

### 3.2. Item Lifecycle Status (product_items.status)

Quy định vòng đời của sản phẩm vật lý:

|  |  |
| --- | --- |
| Value | Meaning |
| IN_STOCK | Đang lưu kho trung tâm. |
| IN_TRANSIT | Đang trên đường vận chuyển. |
| AT_DEALER | Nằm tại cửa hàng bán lẻ, chờ bán. |
| SOLD | Đã bán, nhưng khách chưa kích hoạt sở hữu/bảo hành. |
| REGISTERED | Đã được gắn chủ sở hữu. |
| WARRANTY_ACTIVE | Bảo hành đang có hiệu lực. |
| WARRANTY_CLAIMED | Đang báo lỗi, chờ kỹ thuật viên xử lý. |
| RECALLED | Thuộc diện bị nhà máy gọi thu hồi do lỗi (Block mọi thao tác). |

### 3.3. Warranty Status (warranties.status)

Chuyển đổi một chiều theo BR-14: INACTIVE -> ACTIVE -> CLAIMED -> RESOLVED/REJECTED -> EXPIRED.

|  |  |
| --- | --- |
| Value | Meaning |
| INACTIVE | Cấu hình bảo hành đã tạo nhưng chưa kích hoạt (chờ mua hàng). |
| ACTIVE | Đang trong thời hạn bảo hành hợp lệ. |
| CLAIMED | Khách hàng vừa gửi yêu cầu sửa chữa. |
| RESOLVED | Yêu cầu đã được kỹ thuật viên khắc phục thành công. |
| EXPIRED | Đã quá hạn end_date. |

## 4. Business Rules Derived From Database

Từ thiết kế Schema, các Business Rule sau đây đã được rào chặn chặt chẽ ngay tại tầng Database:

1. Rule Mối quan hệ Batch-Variant (Item Constraint): product_items sử dụng Foreign Key kép (batch_id, variant_id) tham chiếu về batches(id, variant_id). -> Ý nghĩa: Không thể xảy ra lỗi nhầm lẫn lấy một sản phẩm "iPhone" gắn vào lô sản xuất của "Samsung".
2. Rule Bắt buộc có đối tượng truy vết: Cột product_item_id và batch_id trong bảng events không được cùng lúc NULL (CHK constraint). -> Ý nghĩa: Không thể lưu một sự kiện "ma" không gắn với bất kỳ sản phẩm hay lô hàng nào.
3. Rule Cấu trúc Cây (No Self-Loop): Một product_categories không thể nhận chính ID của nó làm parent_id. -> Ý nghĩa: Ngăn chặn lỗi vòng lặp vô tận (Infinite loop) khi load menu danh mục.
4. Rule Dynamic Attributes Validation: Bảng attribute_values giới hạn chỉ cho phép nhập duy nhất 1 kiểu dữ liệu tại 1 thời điểm (value_text, value_number, hoặc value_boolean).
5. Rule Logic Thời gian: * expiry_date >= manufacture_date (Lô hàng).

* ended_at >= owned_at (Chuyển nhượng).
* end_date >= start_date (Bảo hành).

## 5. Missing Documentation & Gap Analysis

Dưới góc nhìn Business Analyst, khi đối chiếu BRD và Schema, có một số điểm khuyết (Gap) cần đội Dev làm rõ:

|  |  |  |
| --- | --- | --- |
| Vấn đề | Mức độ | Lý do & Khuyến nghị |
| Độc quyền sở hữu (BR-03) | HIGH | BRD yêu cầu "Mỗi SP chỉ có 1 chủ sở hữu tại 1 thời điểm". Schema hiện tại phụ thuộc vào code Backend để kiểm tra. Khuyến nghị: Để an toàn tuyệt đối ở tầng DB, nên thêm Partial Unique Index: CREATE UNIQUE INDEX uq_active_ownership ON ownerships (product_item_id) WHERE status = 'ACTIVE'; |
| Cơ chế Sync Vector Search (BR-09) | MEDIUM | BRD dùng Qdrant cho Vector Search. PostgeSQL chỉ đóng vai trò Source of Truth. Schema chưa có cờ báo hiệu (e.g., cờ is_vector_synced trong products) để biết record nào đã được push sang Qdrant thành công. |
| Quy trình Thu hồi (BR-08) | LOW | Schema có status RECALLED ở Batch và Item, nhưng chưa quy định trigger DB nào sẽ tự động cascade update status từ Batch xuống toàn bộ Items thuộc lô đó. (Khả năng team Backend xử lý ở tầng Application). |

## 6. Glossary (Thuật ngữ dự án)

|  |  |
| --- | --- |
| Thuật ngữ | Ý nghĩa trong hệ thống |
| Product | Dòng sản phẩm chung (Ví dụ: Giày Nike Air Force 1). |
| Variant / SKU | Biến thể cụ thể để bán (Ví dụ: Giày Nike AF1 - Màu Trắng - Size 42). |
| Product Item | 1 đôi giày vật lý duy nhất, có in QR Code hoặc Serial Number riêng biệt để truy vết. |
| Batch | Một lô sản xuất gồm nhiều Product Items, ra lò cùng một lúc, có chung hạn sử dụng. |
| Event (Traceability) | Một điểm chạm trong chuỗi cung ứng (VD: Ra khỏi nhà máy, Vào kho đại lý, Được bán ra). Các Event ghép lại tạo thành "Timeline Hành trình". |
| Claim (Warranty) | Yêu cầu báo lỗi/bảo hành từ khách hàng đối với một Product Item. |
| Vector Search | Kỹ thuật tìm kiếm bằng AI trên Qdrant, cho phép khách hàng gõ ngôn ngữ tự nhiên (VD: "tìm đt pin trâu") thay vì tìm chính xác từ khóa. |