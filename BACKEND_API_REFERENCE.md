# Backend API Reference Spec

This document provides a comprehensive specifications and integration guide for the Go Backend APIs of ProductTrace-AI. The specifications are extracted directly from the Go source files, models, routers, and database migrations.

---

## 1. API Summary Table

| Method | Path | Auth | Request Type | Response Type | FE Status | Description |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/public/verify` | Public | Query | `VerifyQRResponse` | Ready | Verify product item QR code scan |
| **POST** | `/api/auth/register` | Public | JSON Body | `RegisterResponse` | Ready | Register a new user account |
| **POST** | `/api/auth/login` | Public | JSON Body | `TokenResponse` | Ready | Authenticate user & get JWT tokens |
| **POST** | `/api/auth/verify-otp` | Public | JSON Body | Empty Success | Ready | Verify user account with OTP |
| **POST** | `/api/auth/resend-otp` | Public | JSON Body | Empty Success | Ready | Resend registration OTP |
| **POST** | `/api/auth/refresh` | Public | JSON Body | `TokenResponse` | Ready | Refresh expired access token |
| **POST** | `/api/auth/logout` | Public | JSON Body | Empty Success | Ready | Invalidate refresh token / Logout |
| **POST** | `/api/auth/forgot-password` | Public | JSON Body | Empty Success | Ready | Request password reset OTP |
| **POST** | `/api/auth/reset-password` | Public | JSON Body | Empty Success | Ready | Reset password using OTP |
| **GET** | `/api/users/profile` | JWT | Header/Query | `UserResponse` | Ready | Get current user or targeted user profile |
| **PUT** | `/api/users/profile/:id` | JWT | JSON Body | `UserResponse` | Ready | Update profile information |
| **PUT** | `/api/users/change-password` | JWT | JSON Body | Empty Success | Ready | Change password for logged-in user |
| **POST** | `/api/users` | Admin | JSON Body | `UserResponse` | Ready | Create a new user (Admin Management) |
| **PUT** | `/api/users/:id` | Admin | JSON Body | `UserResponse` | Ready | Update user (Admin Management) |
| **DELETE** | `/api/users/:id` | Admin | Path Param | Empty Success | Ready | Hard/Soft delete user |
| **GET** | `/api/users` | Admin | Query | `UserListResponse` | Ready | List & search users |
| **GET** | `/api/users/:id` | Admin | Path Param | `UserResponse` | Ready | Get user details |
| **GET** | `/api/batches/:id` | Public | Path Param | `BatchDetailResponse` | Ready | View batch details publicly |
| **GET** | `/api/batches` | JWT | Query | `BatchListResponse` | Ready | Get batch list (with role check for DRAFTs) |
| **GET** | `/api/batches/search` | JWT | Query | `SearchBatchResponse` | Ready | Search batches (approximate query) |
| **GET** | `/api/batches/:id/events` | JWT | Path Param | `[]BatchEventDTO` | Ready | Get history of warehouse/transit events |
| **GET** | `/api/batches/:id/products` | JWT (Admin/Staff/Dealer) | Path Param | `GetBatchProductsResponse` | Ready | View product items inside a batch |
| **GET** | `/api/batches/:id/history` | JWT (Admin/Staff) | Path Param | `GetBatchHistoryResponse` | Ready | View batch modification audit history |
| **POST** | `/api/batches/:id/export` | JWT (Admin/Mgr/Whse) | JSON Body | Empty Success | Ready | Export batch items to a location |
| **GET** | `/api/batches/export-qr/:id` | JWT (Admin/Manuf) | Path Param | Binary PDF | Ready | Download batch QR codes as PDF |
| **POST** | `/api/batches` | JWT (Admin/Manuf) | JSON Body | `BatchCreateResponse` | Ready | Create a new batch & pre-generate items |
| **PATCH** | `/api/batches/:id/status` | JWT (Admin/Manuf) | JSON Body | `BatchStatusResponse` | Ready | Update batch status |
| **DELETE** | `/api/batches/:id` | JWT (Admin/Manuf) | Path Param | Empty Success | Ready | Delete empty batch |
| **GET** | `/api/products` | Public | Query | `ProductListResponse` | Ready | List all products with pagination |
| **GET** | `/api/products/:id` | Public | Path Param | `ProductResponse` | Ready | Get product details by ID |
| **POST** | `/api/products` | JWT (Admin/Manuf) | JSON Body | `ProductResponse` | Ready | Create product and variants |
| **PUT** | `/api/products/:id` | JWT (Admin/Manuf) | JSON Body | `ProductResponse` | Ready | Update product and variants |
| **DELETE** | `/api/products/:id` | JWT (Admin) | Path Param | Empty Success | Ready | Soft-delete product |
| **POST** | `/api/ownership/request-otp` | JWT (Customer) | JSON Body | Empty Success | Ready | Request OTP for ownership registration |
| **POST** | `/api/ownership/register` | JWT (Customer) | JSON Body | `OwnershipDetailRes` | Ready | Confirm OTP & register ownership |
| **POST** | `/api/ownership/admin/request-otp` | JWT (Admin) | JSON Body | Empty Success | Ready | Admin requests OTP for customer registration |
| **POST** | `/api/ownership/admin/register` | JWT (Admin) | JSON Body | `OwnershipDetailRes` | Ready | Admin registers ownership for customer |
| **GET** | `/api/ownership/detail/:product_item_id` | JWT | Path Param | `OwnershipDetailRes` | Ready | View ownership detail history |
| **PUT** | `/api/ownership/:id/transfer` | JWT | JSON Body | Empty Success | Ready | Transfer ownership to another user |
| **DELETE** | `/api/ownership/:id` | JWT | Path Param | Empty Success | Ready | Revoke ownership record |
| **GET** | `/api/ownership` | JWT | Query | `PaginatedOwnershipsRes` | Ready | Search and filter ownership records |
| **POST** | `/api/warranty-claims` | JWT | JSON Body | `WarrantyClaimResponse` | Ready | Create warranty claim |
| **GET** | `/api/locations` | Public | Query | `ListLocationsResponse` | Ready | List locations with search/filters |
| **GET** | `/api/locations/:id` | Public | Path Param | `LocationResponse` | Ready | Get location details |
| **POST** | `/api/locations` | JWT (Admin) | JSON Body | `LocationResponse` | Ready | Create storage/warranty center location |
| **PUT** | `/api/locations/:id` | JWT (Admin) | JSON Body | `LocationResponse` | Ready | Update location attributes |
| **DELETE** | `/api/locations/:id` | JWT (Admin) | Path Param | Empty Success | Ready | Hard delete location |
| **GET** | `/api/dashboard/stats` | JWT (Admin/Staff) | None | `DashboardStats` | Ready | Get system aggregate statistics |
| **GET** | `/api/variants/:id` | Public | Path Param | `VariantResponse` | Ready | Get variant detail by ID |
| **GET** | `/api/variants/product/:product_id` | Public | Query/Path | `ListVariantResponse` | Ready | Get variants of a product |
| **PUT** | `/api/variants/:id` | JWT (Admin/Manuf) | JSON Body | `VariantResponse` | Ready | Update variant details |
| **DELETE** | `/api/variants/:id` | JWT (Admin) | Path Param | Empty Success | Ready | Soft-delete variant |
| **GET** | `/api/attributes` | Public | Query | `ListAttributeResponse` | Ready | List attributes by category |
| **GET** | `/api/attributes/:id` | Public | Path Param | `AttributeResponse` | Ready | Get attribute by ID |
| **POST** | `/api/attributes` | JWT (Admin/Manuf) | JSON Body | `AttributeResponse` | Ready | Create category-specific attribute |
| **PUT** | `/api/attributes/:id` | JWT (Admin/Manuf) | JSON Body | `AttributeResponse` | Ready | Update attribute label/code |
| **DELETE** | `/api/attributes/:id` | JWT (Admin) | Path Param | Empty Success | Ready | Soft-delete attribute |
| **GET** | `/api/attribute-values` | Public | Query | `ListAttributeValueResponse` | Ready | List all attribute values |
| **GET** | `/api/attribute-values/:id` | Public | Path Param | `AttributeValueResponse` | Ready | Get attribute value by ID |
| **PUT** | `/api/attribute-values/:id` | JWT (Admin/Manuf) | JSON Body | `AttributeValueResponse` | Ready | Update attribute value |
| **DELETE** | `/api/attribute-values/:id` | JWT (Admin) | Path Param | Empty Success | Ready | Delete attribute value |
| **GET** | `/api/variants/:id/attributes` | Public | Path Param | `[]AttributeValueResponse` | Ready | Get attribute values for a variant |
| **POST** | `/api/variants/:id/attributes` | JWT (Admin/Manuf) | JSON Body | `[]AttributeValueResponse` | Ready | Bulk assign attributes to a variant |
| **GET** | `/api/product-items` | Public | Query | `ProductItemListResponse` | Ready | List product items |
| **GET** | `/api/product-items/:item_code` | Public | Path Param | `ProductItemDetailDTO` | Ready | Get product item by item code |
| **GET** | `/api/trace/search` | Public | Query | `TraceSearchResponse` | Ready | Trace product item timeline by code/serial |
| **POST** | `/api/trace/export/pdf` | JWT | JSON Body | File/Job | Ready | Export product journey to PDF |
| **POST** | `/api/trace/export/excel` | JWT | JSON Body | File/Job | Ready | Export product journey to Excel |

---

## 2. API Specifications (Detailed Tracing)

### 2.1 Public QR Verification

#### Basic Information
* **HTTP Method**: `GET`
* **Full Path**: `/api/public/verify`
* **Module**: `public`
* **Handler Function**: `VerifyQR`
* **Service Function**: `VerifyQR(ctx, itemCode, token)`
* **Repository Function**: `FindByCodeAndToken(ctx, itemCode, token)`

#### Authentication
* **Required**: No Authentication
* **Required Headers**: None

#### Request Parameters
* **Query Parameters**:
  | Name | Type | Required | Default | Description |
  | :--- | :--- | :--- | :--- | :--- |
  | `item_code` | string | Yes | None | Physical item identifier code (e.g. `PTA-2501-686F493D`) |
  | `token` | string | Yes | None | 32-character hex verification token |

#### Response Spec (200 OK)
```json
{
  "success": true,
  "message": "Item verified successfully",
  "data": {
    "item_code": "PTA-2501-686F493D",
    "serial_number": "SN23650263663452",
    "item_status": "IN_STOCK",
    "scanned_at": "2026-07-13T00:30:00Z",
    "batch": {
      "batch_code": "APL-2026-0001",
      "manufacture_date": "2026-01-01T00:00:00Z",
      "expiry_date": "2027-01-01T00:00:00Z",
      "manufacturer_name": "Apple Vietnam",
      "supplier_name": "FPT Retail",
      "origin_country": "Vietnam",
      "production_place": "Factory A, High-Tech Park, District 9",
      "batch_status": "ACTIVE"
    },
    "product": {
      "product_name": "iPhone 15 Pro",
      "variant_name": "128GB Titanium Gray",
      "variant_sku": "IP15P-128-GRY"
    }
  }
}
```

#### Database Access
* **Read Tables**: `product_items` (joined with `batches`, `product_variants`, `products`)
* **Write Tables**: None
* **Updated Columns**: None
* **Deleted Records**: None

#### Business Logic
Verifies if a physical item exists with the matching `item_code` and `verification_token` combination. It checks if the item is not soft-deleted (`is_deleted = false`) and returns details about the batch production date, origin country, production facility, variant details, and product name.

---

### 2.2 Auth Module

#### POST `/api/auth/register`
* **Handler**: `Register`
* **Service**: `RegisterUser(ctx, email, phone, fullName, password)`
* **Repository**: `Create(ctx, user)`
* **Auth**: None
* **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "phone": "0987654321",
    "full_name": "Nguyen Van A",
    "password": "strongpassword123"
  }
  ```
  * `email` (string, required, validation: `email`)
  * `phone` (string, required)
  * `full_name` (string, required)
  * `password` (string, required, validation: `min=6`)
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Nguyen Van A registered. Verify email via OTP.",
    "data": {
      "full_name": "Nguyen Van A",
      "phone": "0987654321",
      "email": "user@example.com",
      "status": "PENDING"
    }
  }
  ```
* **Database**: Writes into table `users` (status is initialized to `PENDING`).

#### POST `/api/auth/login`
* **Handler**: `Login`
* **Service**: `LoginUser(ctx, email, password)`
* **Auth**: None
* **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "strongpassword123"
  }
  ```
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Logged in successfully",
    "data": {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "d748f293..."
    }
  }
  ```

#### POST `/api/auth/verify-otp`
* **Handler**: `VerifyOTP`
* **Service**: `VerifyOTP(ctx, email, otp)`
* **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "otp": "123456"
  }
  ```
  * `otp` validation: `required,len=6`
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Account verified successfully. You can now log in."
  }
  ```
* **Database**: Updates table `users`, sets `status = 'ACTIVE'`.

#### POST `/api/auth/refresh`
* **Handler**: `RefreshToken`
* **Request Body**:
  ```json
  {
    "refresh_token": "d748f293..."
  }
  ```
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Token refreshed successfully",
    "data": {
      "access_token": "new_access_token...",
      "refresh_token": "new_refresh_token..."
    }
  }
  ```

---

### 2.3 User Module

#### GET `/api/users/profile`
* **Handler**: `GetProfile`
* **Service**: `GetProfile(ctx, userID)`
* **Auth**: JWT Token (Authorization: Bearer <token>)
* **Query Parameters**:
  * `user_id` (string, optional - defaults to token owner)
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Profile loaded successfully",
    "data": {
      "id": "e403d159-28be-40ee-989f-b7a421b369c0",
      "email": "user@example.com",
      "phone": "0987654321",
      "full_name": "Nguyen Van A",
      "role": "CUSTOMER",
      "status": "ACTIVE",
      "avatar": "https://storage.example.com/avatar.jpg",
      "created_at": "2026-07-10T12:00:00Z",
      "updated_at": "2026-07-12T15:00:00Z"
    }
  }
  ```

#### PUT `/api/users/profile/:id`
* **Handler**: `UpdateProfile`
* **Service**: `UpdateProfile(ctx, actorID, targetUserID, req)`
* **Auth**: JWT Token
* **Request Body**:
  ```json
  {
    "full_name": "Nguyen Van B",
    "phone": "0987654322",
    "avatar": "https://storage.example.com/avatar2.jpg"
  }
  ```
* **Database**: Updates columns `full_name`, `phone`, `avatar_url` in table `users`.

---

### 2.4 Batches Module

#### POST `/api/batches`
* **Handler**: `CreateBatch`
* **Service**: `CreateBatch(ctx, req, currentUserID)`
* **Auth**: JWT (ADMIN, MANUFACTURER only)
* **Request Body**:
  ```json
  {
    "variant_id": "8bbdc6d6-f94d-4ba6-993d-d1ef10688a44",
    "prefix": "APL",
    "manufacture_date": "2026-07-13T00:00:00Z",
    "expiry_date": "2027-07-13T00:00:00Z",
    "imported_at": "2026-07-14T00:00:00Z",
    "manufacturer_name": "Foxconn",
    "supplier_name": "Apple Inc",
    "origin_country": "China",
    "production_place": "Shenzhen",
    "quantity": 100
  }
  ```
  * `prefix` validation: `alpha,min=2,max=20`
  * `quantity` validation: `min=0`
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "batch created successfully",
    "data": {
      "id": "e4a77b7f-38fb-476c-843b-2ea1df32d56a",
      "batch_code": "APL-2026-0001",
      "variant_id": "8bbdc6d6-f94d-4ba6-993d-d1ef10688a44",
      "quantity": 100,
      "status": "ACTIVE",
      "created_at": "2026-07-13T00:30:00Z"
    }
  }
  ```
* **Business Logic**: 
  1. Validates that variant exists.
  2. Asserts expiry date >= manufacture date.
  3. Pre-generates the matching quantity of product items (`product_items` table).
  4. Publishes a RabbitMQ event to queue `batch.created`.

#### GET `/api/batches/:id/history`
* **Handler**: `GetBatchHistory`
* **Service**: `GetBatchHistory(ctx, batchID, req, userID)`
* **Auth**: JWT (ADMIN, STAFF only)
* **Query Parameters**:
  * `page` (int, default: 1)
  * `limit` (int, default: 15)
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "get batch history successfully",
    "data": {
      "batchId": "e4a77b7f-38fb-476c-843b-2ea1df32d56a",
      "batchCode": "APL-2026-0001",
      "history": [
        {
          "logId": "23d38734-7548-4ca2-b0de-7a0e10cc74ba",
          "action": "UPDATE",
          "changedFields": {
            "status": {
              "old": "ACTIVE",
              "new": "RECALLED"
            }
          },
          "performedBy": {
            "userId": "d70bc7cd-9f4c-4a3b-82ee-06b23a9d94fa",
            "fullName": "Manager Account",
            "role": "ADMIN"
          },
          "ipAddress": "192.168.1.1",
          "createdAt": "2026-07-13T00:32:00Z"
        }
      ]
    }
  }
  ```

---

### 2.5 Products Module

#### POST `/api/products`
* **Handler**: `CreateProduct`
* **Service**: `CreateProduct(ctx, req, createdBy)`
* **Auth**: JWT (ADMIN, MANUFACTURER only)
* **Request Body**:
  ```json
  {
    "category_id": "4b7b7f32-75fb-476c-843b-2ea1df32d56a",
    "name": "Sony WH-1000XM5",
    "slug": "sony-wh-1000xm5",
    "description": "Premium noise cancelling headphones",
    "thumbnail_url": "https://storage.sony.com/xm5.png",
    "tags": ["audio", "bluetooth", "sony"],
    "metadata": {
      "noise_cancelling_db": 35
    },
    "status": "ACTIVE",
    "variants": [
      {
        "sku": "XM5-BLK",
        "name": "Black Matte",
        "barcode": "4905524021234",
        "price": 8490000,
        "currency": "VND",
        "images": ["https://storage.sony.com/xm5-blk.png"]
      }
    ]
  }
  ```
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Product created successfully",
    "data": {
      "id": "2ba37b7f-38fb-476c-843b-2ea1df32d56a",
      "category_id": "4b7b7f32-75fb-476c-843b-2ea1df32d56a",
      "name": "Sony WH-1000XM5",
      "slug": "sony-wh-1000xm5",
      "description": "Premium noise cancelling headphones",
      "thumbnail_url": "https://storage.sony.com/xm5.png",
      "tags": ["audio", "bluetooth", "sony"],
      "metadata": {
        "noise_cancelling_db": 35
      },
      "status": "ACTIVE",
      "created_by": "d70bc7cd-9f4c-4a3b-82ee-06b23a9d94fa",
      "created_at": "2026-07-13T00:30:00Z",
      "variants": [
        {
          "id": "7ac77b7f-38fb-476c-843b-2ea1df32d56a",
          "sku": "XM5-BLK",
          "name": "Black Matte",
          "price": 8490000,
          "currency": "VND",
          "images": ["https://storage.sony.com/xm5-blk.png"],
          "status": "ACTIVE"
        }
      ]
    }
  }
  ```

---

### 2.6 Ownership Module

#### POST `/api/ownership/request-otp`
* **Handler**: `CustomerRequestOTP`
* **Service**: `CustomerRequestOTP(ctx, req, userID)`
* **Auth**: JWT (CUSTOMER role required)
* **Request Body**:
  ```json
  {
    "qr_code": "PTA-2501-686F493D"
  }
  ```
* **Business Logic**:
  Traces the product item by checking the QR code string. Validates that the item is currently not registered (`status` is not `REGISTERED` or `WARRANTY_ACTIVE`). It then triggers a randomized 6-digit OTP code, saves it to the cache store, and sends an email notification with the OTP verification code to the customer's email (fetched from the active profile JWT claims).

#### POST `/api/ownership/register`
* **Handler**: `CustomerVerifyAndRegister`
* **Service**: `CustomerVerifyAndRegister(ctx, req, userID)`
* **Auth**: JWT (CUSTOMER role required)
* **Request Body**:
  ```json
  {
    "otp": "654321",
    "product_id": "4b7b7f32-75fb-476c-843b-2ea1df32d56a"
  }
  ```
* **Response Body**:
  ```json
  {
    "success": true,
    "message": "Đăng ký quyền sở hữu thành công",
    "data": {
      "ownership_id": "bbdc6d6f-f94d-4ba6-993d-d1ef10688a44",
      "product_id": "4b7b7f32-75fb-476c-843b-2ea1df32d56a",
      "status": "ACTIVE",
      "registration_date": "2026-07-13T00:30:00Z",
      "owner_id": "d70bc7cd-9f4c-4a3b-82ee-06b23a9d94fa",
      "owner_name": "Nguyen Van A",
      "owner_email": "user@example.com",
      "owner_phone": "0987654321",
      "product_name": "Sony WH-1000XM5",
      "product_sku": "XM5-BLK",
      "ownership_history": []
    }
  }
  ```

---

### 2.7 Warranty Claim Module

#### POST `/api/warranty-claims`
* **Handler**: `CreateWarrantyClaim`
* **Service**: `CreateWarrantyClaim(ctx, userID, req)`
* **Auth**: JWT (All authenticated users)
* **Request Headers**:
  * `Authorization`: `Bearer xxx`
  * `X-User-ID`: `<uuid>` (Can also be set downstream by middleware)
* **Request Body**:
  ```json
  {
    "product_id": "5cb2cd59-67ee-45ba-aa01-d70fc891f93d",
    "issue_title": "Loa bên trái không phát ra âm thanh",
    "issue_description": "Sau khi nghe nhạc khoảng 10 phút, tai nghe bên trái bị rè rồi tắt hẳn.",
    "contact_phone": "0987654321",
    "contact_email": "customer@example.com",
    "preferred_service_center": "d70bc7cd-9f4c-4a3b-82ee-06b23a9d94fa",
    "attachments": ["https://storage.example.com/issue-video.mp4"]
  }
  ```
* **Response Body**:
  ```json
  {
    "message": "Warranty Claim Created",
    "data": {
      "id": "e44d159f-d38e-4a62-b912-70fc89115fa0",
      "claim_number": "CLAIM-20260713-0001",
      "product_id": "5cb2cd59-67ee-45ba-aa01-d70fc891f93d",
      "issue_title": "Loa bên trái không phát ra âm thanh",
      "issue_description": "Sau khi nghe nhạc khoảng 10 phút, tai nghe bên trái bị rè rồi tắt hẳn.",
      "contact_phone": "0987654321",
      "contact_email": "customer@example.com",
      "preferred_service_center_id": "d70bc7cd-9f4c-4a3b-82ee-06b23a9d94fa",
      "attachments": ["https://storage.example.com/issue-video.mp4"],
      "status": "OPEN",
      "created_at": "2026-07-13T00:30:00Z",
      "updated_at": "2026-07-13T00:30:00Z"
    }
  }
  ```
* **Database Writes**:
  * Reads `ownerships` to verify current active ownership of the customer (`owner_id`).
  * Reads `warranties` to verify active status and validity duration.
  * Writes record in `warranty_claims` (default status: `OPEN`).

---

## 3. Frontend Inconsistencies & Integrations TODO

### 3.1 Casing Discrepancies
* **Snake Case vs Camel Case**:
  * Public APIs, user profile, batches list, and ownership endpoints use **`snake_case`** keys (e.g. `manufacture_date`, `expiry_date`, `variant_id`, `batch_code`).
  * Search Batches (`GET /api/batches/search`), Batch Products (`GET /api/batches/:id/products`), Batch History (`GET /api/batches/:id/history`), and Trace Search (`GET /api/trace/search`) responses use **`camelCase`** keys (e.g., `batchId`, `batchCode`, `productName`, `manufacturingDate`, `totalRecords`, `currentPage`).
  * **Frontend Impact**: State mappers must normalize payload properties depending on the specific endpoint called.

### 3.2 Header Inconsistencies
* **`X-User-ID` vs `X-User-Id`**:
  * In the auth middleware, the user ID is placed in request headers using `c.Request.Header.Set("X-User-Id", ...)`.
  * The user handler fetches it via `c.GetHeader("X-User-Id")`.
  * The warranty claim handler retrieves it via `c.GetHeader("X-User-ID")`.
  * **Frontend Recommendation**: Ensure that when calling backend endpoints directly (bypassing Kong proxy or Gateway headers), frontend axios interceptors attach standard `Authorization: Bearer <token>` token, and downstream routers handle context retrieval safely.

### 3.3 Response wrappers
* The backend exposes a wrapper: `{ success: boolean, message: string, data: any }`.
* In `warranty_claim_handler.go`, the created response is:
  ```json
  {
    "message": "Warranty Claim Created",
    "data": { ... }
  }
  ```
  Note that the standard `"success": true` key is missing in this response!
  * **Frontend Impact**: Frontend axios hooks should not strictly rely on `response.data.success` to resolve query status, but rather inspect the HTTP status code range `2xx`.

---

## 4. Integration Report

### 4.1 Ready APIs
* All core flows for anonymous QR verification, registration, OTP validation, batch creation/search, and product list tracking are fully implemented and verified against repository tests.

### 4.2 Need Frontend Adjustments
* **CamelCase mapping**: Implement mapper functions in frontend services when dealing with trace searches, batch historical logs, and batch product lists.
* **Axios Response Interceptors**: Use response headers/HTTP status codes rather than looking up `.success` boolean directly in JSON payloads due to inconsistency in `POST /api/warranty-claims`.

### 4.3 Recommended Refactors
* Standardize all JSON tags across modules to use consistent `snake_case`.
* Unify context header retrievals so that user identity is always read from Gin's local context (`c.Get("user_id")`) instead of parsing string header canonicals (`c.GetHeader("X-User-Id")`).
