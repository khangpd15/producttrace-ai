# API Specification & Verification Plan (Batch & Trace Modules)

This document provides a comprehensive API specification and test cases matrix for testing the features of the **Batch** and **Traceability (Trace)** modules in **ProductTrace AI**.

---

## 1. Authentication Flow

Most endpoints (except anonymous trace search and public batch view) require an `Authorization` header with a bearer token.

### Login Endpoint
* **Endpoint**: `POST /api/auth/login`
* **Content-Type**: `application/json`
* **Request Body**:
  ```json
  {
    "email": "admin@producttrace.com",
    "password": "Password123!"
  }
  ```
* **Response (Success)**:
  ```json
  {
    "success": true,
    "message": "Login successfully",
    "data": {
      "accessToken": "eyJhbGciOi...",
      "refreshToken": "eyJhbGciOi...",
      "user": {
        "id": "7c9e66b3-e5d8-4f10-bf91-b3b35520ee11",
        "email": "admin@producttrace.com",
        "role": "ADMIN"
      }
    }
  }
  ```

*To authenticate subsequent requests, include:*
```http
Authorization: Bearer <accessToken>
```

---

## 2. Batch Module API Specification

### 2.1 Get Batch Detail (Public QR Scan)
* **Endpoint**: `GET /api/batches/:id` (where `:id` is the `batchCode`, e.g., `APL-2026-0001`)
* **Auth Required**: No (Public)
* **Response (Success)**:
  ```json
  {
    "success": true,
    "data": {
      "id": "7c9e66b3-e5d8-4f10-bf91-b3b35520ee11",
      "batchCode": "APL-2026-0001",
      "status": "ACTIVE",
      "quantity": 100,
      "manufactureDate": "2026-07-01T00:00:00Z",
      "expiryDate": "2027-07-01T00:00:00Z",
      "productName": "Red Apple Premium"
    }
  }
  ```

---

### 2.2 Get Batch List with Filter
* **Endpoint**: `GET /api/batches`
* **Auth Required**: Yes (`ADMIN`, `STAFF`, `DEALER`, `MANUFACTURER`, `WAREHOUSE`, `MANAGER`)
* **Query Parameters**:
  - `status`: Filter by status (`DRAFT`, `ACTIVE`, `EXPIRED`, `RECALLED`, `LOCKED` or `ALL`)
  - `page`: Page index (default: `1`)
  - `limit`: Page limit (default: `10`)
* **Role Restricting Rules**:
  - `DRAFT` status filter is only allowed for `ADMIN` role. Any other role requesting `status=DRAFT` gets `403 Forbidden`.
  - Non-Admin roles listing `status=ALL` or omitting status will not see `DRAFT` batches in the result list (filtered out automatically).

---

### 2.3 Search Batches
* **Endpoint**: `GET /api/batches/search`
* **Auth Required**: Yes (All authenticated users)
* **Query Parameters**:
  - `keyword`: Search query (Max length: `100` characters. Longer keywords get `400 Bad Request`)
  - `sortField`: Sort by field (default: `created_at`)
  - `sortOrder`: `ASC` or `DESC`

---

### 2.4 Create Batch
* **Endpoint**: `POST /api/batches`
* **Auth Required**: Yes (`ADMIN`, `MANUFACTURER` only)
* **Request Body**:
  ```json
  {
    "variantId": "6c9e66b3-e5d8-4f10-bf91-b3b35520ee22",
    "prefix": "APL",
    "manufactureDate": "2026-07-01T00:00:00Z",
    "expiryDate": "2027-07-01T00:00:00Z",
    "quantity": 100,
    "manufacturerName": "Apple Farm Inc",
    "supplierName": "Farm Distrib",
    "originCountry": "Vietnam",
    "productionPlace": "Lam Dong"
  }
  ```
* **Validation Constraints**:
  - `quantity` must be $> 0$.
  - `manufactureDate` must not be in the future.
  - `expiryDate` must be $\ge$ `manufactureDate`.

---

### 2.5 Update Batch Status
* **Endpoint**: `PATCH /api/batches/:id/status`
* **Auth Required**: Yes (`ADMIN`, `MANUFACTURER` only)
* **Request Body**:
  ```json
  {
    "status": "EXPIRED"
  }
  ```
* **Valid Enums**: `ACTIVE`, `EXPIRED`, `RECALLED`, `BLOCKED`

---

### 2.6 Delete Batch
* **Endpoint**: `DELETE /api/batches/:id`
* **Auth Required**: Yes (`ADMIN`, `MANUFACTURER` only)
* **Constraints**:
  - Cannot delete if the batch has linked product items or events.

---

### 2.7 Get Batch Products
* **Endpoint**: `GET /api/batches/:id/products`
* **Auth Required**: Yes (`ADMIN`, `STAFF`, `DEALER`)
* **Query Parameters**:
  - `status`: Optional filter status
  - `page`: Page index
  - `limit`: Limit per page (Required)

---

### 2.8 Get Batch History
* **Endpoint**: `GET /api/batches/:id/history`
* **Auth Required**: Yes (`ADMIN`, `STAFF` only)

---

## 3. Traceability Module API Specification

### 3.1 Trace Search Timeline (Public & Protected Hybrid)
* **Endpoint**: `GET /api/trace/search`
* **Auth Required**: No (Public)
* **Query Parameters**:
  - `code`: Product code or Serial number (e.g., `PT-MILK-SN0001`) (Required, $3$ to $100$ characters)
  - `fromDate`: ISO-8601 start date, e.g., `2026-07-01T00:00:00Z` (Optional)
  - `toDate`: ISO-8601 end date, e.g., `2026-07-08T00:00:00Z` (Optional)
  - `eventTypes`: Comma-separated enums, e.g., `PRODUCED,SALE` (Optional)
* **Behavior Rules**:
  - **Rate Limiting**: Limited to 30 requests/minute per client IP. Returns `429 Too Many Requests` if exceeded.
  - **Information Shielding**: Non-privileged users (Public or role `CUSTOMER`) are shielded from internal warehouse/transit events (`PACKED`, `WAREHOUSE_IN`, `WAREHOUSE_OUT`, `TRANSPORTED`). Only public events (`PRODUCED`, `SALE`, `REGISTERED`, `WARRANTY_CLAIM`, `WARRANTY_RESOLVED`, `RECALLED`) are returned.
  - **Recalled Warnings**: Returns a Warning string if product status or batch is `RECALLED`.

---

### 3.2 Export PDF
* **Endpoint**: `POST /api/trace/export/pdf`
* **Auth Required**: Yes (`ADMIN`, `STAFF`, `DEALER`)
* **Request Body**:
  ```json
  {
    "productItemId": "c20befe6-e4b9-4f37-b514-e66233ef04a1",
    "theme": "CLASSIC_NAVY",
    "includeAuditLogs": true
  }
  ```
* **Output**:
  - If timeline has $< 10$ events: streams binary file directly (`200 OK`).
  - If timeline has $\ge 10$ events: returns a background job ID (`202 Accepted`).

---

### 3.3 Export Excel
* **Endpoint**: `POST /api/trace/export/excel`
* **Auth Required**: Yes (`ADMIN`, `STAFF`, `DEALER`)
* **Request Body**:
  ```json
  {
    "productItemId": "c20befe6-e4b9-4f37-b514-e66233ef04a1",
    "batchId": "6f2bc881-8b21-4f10-9111-a887b2210a12",
    "fromDate": "2026-07-01T00:00:00Z",
    "toDate": "2026-07-08T23:59:59Z"
  }
  ```
* **Output**:
  - If single item export with $< 10$ events: streams binary file directly (`200 OK`).
  - Otherwise (large dataset or batch level): returns a background job ID (`202 Accepted`).

---

## 4. Test Verification Matrix (Verification Plan)

| Test ID | Module | Endpoint | Method | Query / Body Payload | Auth / Role | Expected Outcome |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **TC-B-001** | Batch | `/api/batches` | `GET` | `status=DRAFT` | Admin | `200 OK` (list of draft batches) |
| **TC-B-002** | Batch | `/api/batches` | `GET` | `status=DRAFT` | Customer / Staff | `403 Forbidden` |
| **TC-B-003** | Batch | `/api/batches/search` | `GET` | `keyword=<105 chars>` | Admin | `400 Bad Request` (too long) |
| **TC-B-004** | Batch | `/api/batches` | `POST` | ExpiryDate before ManufactureDate | Admin | `400 Bad Request` / `422 Unprocessable` |
| **TC-B-005** | Batch | `/api/batches` | `POST` | ManufactureDate in future | Admin | `400 Bad Request` / `422 Unprocessable` |
| **TC-B-006** | Batch | `/api/batches/:id` | `DELETE` | Batch with active items | Admin | `400 Bad Request` (Linked items check) |
| **TC-T-001** | Trace | `/api/trace/search` | `GET` | `code=PT-MILK-SN0001` | Public | `200 OK` (no internal warehouse events) |
| **TC-T-002** | Trace | `/api/trace/search` | `GET` | `code=PT-MILK-SN0001` | Admin (Token) | `200 OK` (all warehouse events included) |
| **TC-T-003** | Trace | `/api/trace/search` | `GET` | `code=PT-MILK-SN0001` (35 times/min) | Public | `429 Too Many Requests` (Rate Limited) |
| **TC-T-004** | Trace | `/api/trace/export/pdf` | `POST` | `{"productItemId": "...", "theme": "CLASSIC_NAVY"}` (Timeline < 10) | Admin | `200 OK` (streams PDF binary) |
| **TC-T-005** | Trace | `/api/trace/export/pdf` | `POST` | `{"productItemId": "...", "theme": "CLASSIC_NAVY"}` (Timeline >= 10) | Admin | `202 Accepted` (JSON containing `jobId`) |
