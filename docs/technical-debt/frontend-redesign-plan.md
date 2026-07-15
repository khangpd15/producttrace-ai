# KẾ HOẠCH THIẾT KẾ LẠI FRONTEND - PRODUCTTRACE-AI

**Người lập**: Solution Architect  
**Ngày**: 15/07/2026  
**Căn cứ**: `docs/project-audit.md`  
**Stack hiện tại**: Vite + React 19 + TypeScript + TailwindCSS + React Router v7

---

## TỔNG QUAN ƯU TIÊN

Dựa trên audit, Frontend hiện chỉ có 3 trang (Login mock, ForgotPassword, ResetPassword).  
Toàn bộ nghiệp vụ lõi chưa được xây dựng.

**Thứ tự ưu tiên xử lý:**

| Ưu tiên | Trang / Module | Lý do |
|:---:|:---|:---|
| 🔴 P0 | Login Page | Entry point bị mock cứng, block toàn bộ luồng |
| 🔴 P0 | Hạ tầng chung (API Layer, Auth Context, Router Guard) | Prerequisite của mọi trang khác |
| 🟠 P1 | Admin Dashboard | Trang chủ sau login của ADMIN/STAFF |
| 🟠 P1 | Product Management | Backend hoàn chỉnh, FE chưa có |
| 🟠 P1 | Customer Portal (Dashboard + Ownership) | Luồng nghiệp vụ cốt lõi cho khách hàng |
| 🟡 P2 | QR Scanner & Timeline | Giá trị demo cao |
| 🟡 P2 | AI Hybrid Search | Backend hoàn chỉnh, FE chưa có |
| 🟡 P2 | Batch & Trace Management | Backend hoàn chỉnh, FE chưa có |
| 🟢 P3 | Warranty Management | Backend thiếu GET API, cần phối hợp BE |
| 🟢 P3 | Location & Dealer Map | Backend hoàn chỉnh |
| 🟢 P3 | User Management | Admin-only, ít urgency |

---

## PHẦN I — HẠ TẦNG CHUNG (Infrastructure)

> Phải hoàn thành trước khi xây dựng bất kỳ trang nào.

---

### INF-1: Cấu trúc thư mục mới

**Mục tiêu**: Tổ chức lại `src/` theo kiến trúc Feature-based, dễ scale.

**Trạng thái hiện tại**: Chỉ có `src/pages/` với 3 file đơn lẻ. Không có layers API, hooks, context.

**Cấu trúc đề xuất**:
```
src/
├── api/               # Axios instances + API functions theo domain
│   ├── client.ts      # Axios base client (Go Core + Nest AI)
│   ├── auth.api.ts
│   ├── product.api.ts
│   ├── batch.api.ts
│   ├── ownership.api.ts
│   ├── warranty.api.ts
│   ├── location.api.ts
│   ├── dashboard.api.ts
│   ├── trace.api.ts
│   └── search.api.ts
├── components/        # Shared UI components
│   ├── ui/            # Atoms: Button, Input, Badge, Modal, Table, Spinner
│   ├── layout/        # AppShell, Sidebar, Topbar, PageHeader
│   └── shared/        # DataTable, SearchBar, StatusBadge, ConfirmDialog
├── context/
│   └── AuthContext.tsx # JWT decode, user role, login/logout
├── hooks/             # Custom hooks per domain
│   ├── useAuth.ts
│   ├── useProducts.ts
│   └── ...
├── pages/             # Page components theo role
│   ├── auth/
│   ├── admin/
│   ├── customer/
│   └── public/
├── routes/
│   └── index.tsx      # Centralized router với ProtectedRoute + role guard
├── types/             # TypeScript interfaces từ BE schemas
│   ├── auth.types.ts
│   ├── product.types.ts
│   └── ...
└── utils/
    ├── token.ts       # JWT decode/store helper
    └── format.ts      # Date, currency formatters
```

**Việc cần làm**:
- [ ] Tạo cấu trúc thư mục trên
- [ ] Cài thêm: `axios`, `@tanstack/react-query`, `react-hot-toast`, `date-fns`
- [ ] Tùy chọn: `recharts` (charts Dashboard), `html5-qrcode` (QR Scanner)

---

### INF-2: API Client Layer

**Mục tiêu**: Tập trung toàn bộ HTTP calls, xử lý JWT interceptor tự động.

**API cần dùng**:
- Go Core Service: `http://localhost:8080/api`
- Nest AI Service: `http://localhost:3000`

**Component cần tạo**: `src/api/client.ts`

**Trạng thái hiện tại**: Không có. ForgotPasswordPage dùng `fetch()` thẳng trong component.

**Việc cần làm**:
- [ ] Tạo Axios instance cho Go Core với interceptor tự động gắn `Bearer token`
- [ ] Tạo Axios instance riêng cho Nest AI Service
- [ ] Xử lý `401 Unauthorized` → tự động logout và redirect về `/login`
- [ ] Cấu hình `.env` với `VITE_GO_CORE_URL` và `VITE_NEST_AI_URL`

---

### INF-3: Auth Context & Protected Route

**Mục tiêu**: Quản lý trạng thái xác thực toàn cục, phân quyền theo role.

**API cần dùng**: `POST /api/auth/login` (Go Core)

**Component cần tạo**:
- `src/context/AuthContext.tsx` — cung cấp `user`, `token`, `login()`, `logout()`
- `src/routes/ProtectedRoute.tsx` — kiểm tra auth + role trước khi render

**Trạng thái hiện tại**: Không có. LoginPage chỉ `console.log`.

**Việc cần làm**:
- [ ] Lưu JWT vào `localStorage` sau khi login thành công
- [ ] Decode JWT payload để lấy `role` (ADMIN, STAFF, DEALER, CUSTOMER)
- [ ] ProtectedRoute redirect về `/login` nếu chưa auth
- [ ] ProtectedRoute hiện `403` nếu đúng auth nhưng sai role

---

## PHẦN II — TRANG ƯU TIÊN CAO (P0 / P1)

---

### PAGE-1: Login Page (🔴 P0 — Mock)

**Mục tiêu**: Thay thế mock `console.log` bằng luồng đăng nhập thực tế với Go Core Service, điều hướng theo role sau khi đăng nhập.

**API cần dùng**:
```
POST /api/auth/login
Body: { email, password }
Response: { access_token, refresh_token, user: { id, name, email, role } }
```

**Component cần tạo**:
- Redesign toàn bộ UI `LoginPage.tsx` (hiện đang dùng class Tailwind thô, không có design system)
- `src/components/ui/Input.tsx` — reusable input với error state
- `src/components/ui/Button.tsx` — reusable button với loading state

**Trạng thái hiện tại**: 
- UI có sẵn nhưng `handleSubmit` chỉ `console.log`
- Không có error handling, loading state, hay redirect

**Việc cần làm**:
- [ ] Gọi `POST /api/auth/login` thực tế
- [ ] Lưu token vào AuthContext
- [ ] Redirect sau login theo role:
  - `ADMIN/STAFF` → `/admin/dashboard`
  - `DEALER` → `/dealer/dashboard`
  - `CUSTOMER` → `/customer/dashboard`
- [ ] Hiện thông báo lỗi khi sai credentials
- [ ] Thêm loading spinner lúc đang call API
- [ ] Redesign UI: dark mode, gradient brand, animation

---

### PAGE-2: Admin Dashboard (🟠 P1 — Chưa có)

**Mục tiêu**: Trang tổng quan cho ADMIN/STAFF hiển thị số liệu thực tế từ Backend.

**API cần dùng**:
```
GET /api/dashboard/stats         → Tổng số sản phẩm, batches, users, owners
GET /api/dashboard/activities    → Hoạt động gần đây
GET /api/dashboard/alerts        → Cảnh báo hệ thống
GET /api/dashboard/sales-chart   → Dữ liệu biểu đồ bán hàng
```

**Component cần tạo**:
- `src/components/layout/AdminShell.tsx` — Sidebar + Topbar cho admin
- `src/components/layout/Sidebar.tsx` — Navigation menu theo role
- `src/pages/admin/DashboardPage.tsx` — Trang dashboard chính
- `src/components/shared/StatCard.tsx` — Card hiển thị số liệu KPI
- `src/components/shared/ActivityFeed.tsx` — Danh sách hoạt động gần đây
- `src/components/shared/SalesChart.tsx` — Biểu đồ (dùng Recharts)

**Trạng thái hiện tại**: Chưa có. Backend hoàn chỉnh.

**Việc cần làm**:
- [ ] Tạo AdminShell layout với Sidebar collapsible
- [ ] Fetch và hiển thị 4 stat cards (Products, Batches, Users, Owners)
- [ ] Hiển thị activity feed với real-time-like polling (mỗi 30s)
- [ ] Tích hợp biểu đồ bán hàng dạng line chart
- [ ] Hiển thị danh sách alerts với badge màu theo severity

---

### PAGE-3: Product Management (🟠 P1 — Chưa có)

**Mục tiêu**: CRUD đầy đủ sản phẩm, danh mục, biến thể cho ADMIN/STAFF.

**API cần dùng**:
```
# Products
GET    /api/products             → Danh sách (có phân trang, filter)
POST   /api/products             → Tạo sản phẩm mới
GET    /api/products/:id         → Chi tiết
PUT    /api/products/:id         → Cập nhật
DELETE /api/products/:id         → Xóa

# Categories
GET    /api/product-categories   → Cây danh mục
POST   /api/product-categories   → Tạo danh mục
PUT    /api/product-categories/:id
DELETE /api/product-categories/:id

# Variants
GET    /api/product-variants     → Danh sách biến thể của sản phẩm
POST   /api/product-variants
PUT    /api/product-variants/:id
DELETE /api/product-variants/:id

# Product Items (QR instances)
GET    /api/product-items        → Danh sách sản phẩm vật lý (QR codes)
GET    /api/product-items/:id
```

**Component cần tạo**:
- `src/pages/admin/products/ProductListPage.tsx`
- `src/pages/admin/products/ProductDetailPage.tsx`
- `src/pages/admin/products/ProductFormPage.tsx` (Create/Edit)
- `src/pages/admin/categories/CategoryPage.tsx` (tree view)
- `src/components/shared/DataTable.tsx` — Table chung có sort, filter, pagination
- `src/components/shared/ProductCard.tsx`
- `src/components/shared/CategoryTree.tsx`

**Trạng thái hiện tại**: Chưa có trang nào. Backend hoàn chỉnh đầy đủ CRUD.

**Việc cần làm**:
- [ ] Tạo DataTable component tái sử dụng (pagination server-side)
- [ ] Product list với search, filter theo category
- [ ] Product form: multi-step (Info → Variants → Attributes)
- [ ] Category tree CRUD
- [ ] Product items list với QR code preview

---

### PAGE-4: Customer Dashboard & Ownership (🟠 P1 — Chưa có)

**Mục tiêu**: Customer Portal cho phép khách hàng xem sản phẩm sở hữu và đăng ký sở hữu mới.

**API cần dùng**:
```
# Ownership
GET    /api/ownerships/my              → Danh sách sản phẩm của tôi
GET    /api/ownerships/:id             → Chi tiết ownership
POST   /api/ownerships/register        → Đăng ký sở hữu (gửi OTP)
POST   /api/ownerships/verify-otp      → Xác thực OTP để hoàn tất đăng ký
POST   /api/ownerships/:id/transfer    → Chuyển nhượng sở hữu

# Product Items (tra cứu theo QR code)
GET    /api/product-items/:code        → Tra cứu sản phẩm bằng mã QR
```

**Component cần tạo**:
- `src/components/layout/CustomerShell.tsx` — Shell riêng cho Customer
- `src/pages/customer/DashboardPage.tsx` — Danh sách sản phẩm đang sở hữu
- `src/pages/customer/OwnershipDetailPage.tsx` — Chi tiết 1 sản phẩm
- `src/pages/customer/RegisterOwnershipPage.tsx` — Form đăng ký sở hữu + OTP
- `src/pages/customer/TransferOwnershipPage.tsx` — Form chuyển nhượng
- `src/components/shared/OtpInput.tsx` — OTP input 6 digits
- `src/components/shared/ProductOwnershipCard.tsx`

**Trạng thái hiện tại**: Chưa có. Backend hoàn chỉnh.

**Việc cần làm**:
- [ ] Customer shell layout (khác Admin, nhẹ hơn)
- [ ] Dashboard hiển thị grid sản phẩm đang sở hữu
- [ ] Luồng đăng ký ownership: nhập mã QR → gửi OTP → nhập OTP → xác nhận
- [ ] Luồng chuyển nhượng: chọn sản phẩm → nhập email người nhận → xác nhận

---

## PHẦN III — TRANG ƯU TIÊN TRUNG (P2)

---

### PAGE-5: QR Scanner & Product Timeline (🟡 P2 — Chưa có)

**Mục tiêu**: Trang public cho phép bất kỳ ai quét QR để xem lịch sử hành trình sản phẩm.

**API cần dùng**:
```
GET /public/verify?code=PTA-YYMM-XXXXXXXX   → Thông tin QR public (Go Core)
GET /api/traces/:productItemCode             → Timeline đầy đủ (Go Core, auth)
GET /api/traces/:code/export?format=pdf      → Xuất PDF
GET /api/traces/:code/export?format=excel    → Xuất Excel
```

**Component cần tạo**:
- `src/pages/public/QRScanPage.tsx` — Camera + manual input
- `src/pages/public/ProductTimelinePage.tsx` — Timeline visualizer
- `src/components/shared/QRScanner.tsx` — Wrapper html5-qrcode
- `src/components/shared/Timeline.tsx` — Vertical timeline component
- `src/components/shared/EventCard.tsx` — Card mỗi sự kiện trong timeline
- `src/components/shared/ExportButtons.tsx` — PDF/Excel download

**Trạng thái hiện tại**: Chưa có. Backend hoàn chỉnh (verify + export).

**Việc cần làm**:
- [ ] Trang public `/scan` với camera QR scanner
- [ ] Fallback: nhập mã QR thủ công
- [ ] Hiển thị timeline dạng vertical stepper với icon sự kiện
- [ ] Map preview vị trí (nếu có tọa độ geo)
- [ ] Nút export PDF/Excel

---

### PAGE-6: AI Hybrid Search (🟡 P2 — Chưa có)

**Mục tiêu**: Giao diện tìm kiếm sản phẩm thông minh dùng vector search của Nest AI.

**API cần dùng**:
```
POST /search/hybrid   (Nest AI Service)
Body: {
  query: string,
  filters: { category?: string, manufacturer?: string, province?: string },
  limit: number,
  offset: number
}
Response: { results: [...], total, scores }
```

**Component cần tạo**:
- `src/pages/search/AISearchPage.tsx` — Trang tìm kiếm chính
- `src/components/shared/SearchBar.tsx` — Input với debounce 300ms
- `src/components/shared/FilterPanel.tsx` — Sidebar filter: category, province
- `src/components/shared/SearchResultCard.tsx` — Hiển thị sản phẩm + relevance score
- `src/components/shared/RelevanceScore.tsx` — Visual score indicator

**Trạng thái hiện tại**: Chưa có. Nest AI backend hoàn chỉnh (hybrid search + RRF ranking).

**Việc cần làm**:
- [ ] Search page accessible tại `/search`
- [ ] Debounce input 300ms trước khi gọi API
- [ ] Filter panel với multiselect cho category, province
- [ ] Hiển thị kết quả với relevance score badge
- [ ] Empty state và loading skeleton
- [ ] Tích hợp vào Topbar (global search)

---

### PAGE-7: Batch & Product Item Management (🟡 P2 — Chưa có)

**Mục tiêu**: Quản lý lô hàng, sinh QR codes hàng loạt.

**API cần dùng**:
```
GET    /api/batches              → Danh sách lô hàng
POST   /api/batches              → Tạo lô hàng mới
GET    /api/batches/:id          → Chi tiết + danh sách product items
GET    /api/batches/:id/export   → Xuất file QR PDF
DELETE /api/batches/:id
```

**Component cần tạo**:
- `src/pages/admin/batches/BatchListPage.tsx`
- `src/pages/admin/batches/BatchDetailPage.tsx`
- `src/pages/admin/batches/BatchFormPage.tsx`
- `src/components/shared/QRCodeGrid.tsx` — Grid hiển thị QR codes
- `src/components/shared/BatchStatusBadge.tsx`

**Trạng thái hiện tại**: Chưa có. Backend hỗ trợ tạo lô, xuất QR PDF.

**Việc cần làm**:
- [ ] Batch list với filter theo trạng thái, ngày tạo
- [ ] Form tạo batch: chọn product variant, nhập số lượng
- [ ] Batch detail: xem danh sách QR codes đã sinh
- [ ] Nút "Xuất QR PDF" download trực tiếp

---

## PHẦN IV — TRANG ƯU TIÊN THẤP (P3)

---

### PAGE-8: Warranty Management (🟢 P3 — Thiếu Backend GET API)

**Mục tiêu**: Khách hàng gửi yêu cầu bảo hành và xem lịch sử.

**API cần dùng**:
```
POST /api/warranty-claims               → Tạo yêu cầu (Backend có)
GET  /api/warranty-claims/my            → Lịch sử bảo hành (⚠ Backend CHƯA CÓ)
GET  /api/warranty-claims/:id           → Chi tiết claim (⚠ Backend CHƯA CÓ)
PUT  /api/warranty-claims/:id/status    → Cập nhật trạng thái (⚠ Backend CHƯA CÓ)
```

**Component cần tạo**:
- `src/pages/customer/WarrantyPage.tsx`
- `src/pages/customer/WarrantyFormPage.tsx` — Form gửi yêu cầu
- `src/components/shared/WarrantyStatusBadge.tsx`

**Trạng thái hiện tại**: 
- Backend chỉ có `POST /api/warranty-claims`
- Service dùng mock adapter cho Event/Audit/Notification
- Hoàn toàn chưa có FE

**⚠ Dependency**: Cần Backend thêm GET APIs trước khi FE có thể hoàn thiện.

**Việc cần làm**:
- [ ] **[Backend trước]** Thêm `GET /api/warranty-claims/my` và `GET /api/warranty-claims/:id`
- [ ] Implement form gửi yêu cầu bảo hành (chỉ cần POST — có thể làm ngay)
- [ ] Warranty history list (chờ BE)
- [ ] Warranty detail với trạng thái (chờ BE)

---

### PAGE-9: Location & Dealer Map (🟢 P3 — Chưa có)

**Mục tiêu**: Hiển thị bản đồ kho hàng, đại lý, trung tâm bảo hành.

**API cần dùng**:
```
GET /api/locations               → Danh sách địa điểm
GET /api/locations/:id           → Chi tiết
POST /api/locations              → Tạo địa điểm (Admin)
PUT  /api/locations/:id          → Cập nhật (Admin)
```

**Component cần tạo**:
- `src/pages/admin/locations/LocationListPage.tsx`
- `src/pages/public/DealerMapPage.tsx` — Bản đồ công khai
- `src/components/shared/LocationCard.tsx`
- `src/components/shared/MapView.tsx` — Wrapper Leaflet/Google Maps

**Trạng thái hiện tại**: Chưa có. Backend hoàn chỉnh với PostGIS.

**Việc cần làm**:
- [ ] Cài Leaflet.js hoặc dùng Google Maps Embed
- [ ] Hiển thị pins trên bản đồ theo loại (kho, đại lý, bảo hành)
- [ ] Admin CRUD locations với picker tọa độ
- [ ] Trang public `/dealers` không cần auth

---

### PAGE-10: User Management (🟢 P3 — Chưa có)

**Mục tiêu**: ADMIN quản lý toàn bộ người dùng, phân quyền, khóa tài khoản.

**API cần dùng**:
```
GET    /api/users               → Danh sách users (có phân trang)
GET    /api/users/:id           → Chi tiết user
PUT    /api/users/:id           → Cập nhật thông tin
PUT    /api/users/:id/lock      → Khóa tài khoản
PUT    /api/users/:id/unlock    → Mở khóa tài khoản
DELETE /api/users/:id           → Xóa user
```

**Component cần tạo**:
- `src/pages/admin/users/UserListPage.tsx`
- `src/pages/admin/users/UserDetailPage.tsx`
- `src/components/shared/RoleBadge.tsx`
- `src/components/shared/UserAvatar.tsx`

**Trạng thái hiện tại**: Chưa có. Backend hoàn chỉnh.

**Việc cần làm**:
- [ ] User list với filter theo role, trạng thái
- [ ] User detail: xem thông tin, lịch sử sở hữu
- [ ] Action: lock/unlock account với confirm dialog
- [ ] Role badge: ADMIN / STAFF / DEALER / CUSTOMER với màu phân biệt

---

## PHẦN V — TRANG AUTH CẦN NÂNG CẤP

---

### PAGE-11: ForgotPassword & ResetPassword (Nâng cấp UI)

**Mục tiêu**: Giữ nguyên logic đã kết nối, nâng cấp UI phù hợp design system mới.

**Trạng thái hiện tại**: Logic kết nối với Nest AI Service đã hoàn chỉnh. Chỉ cần redesign UI.

**Việc cần làm**:
- [ ] Áp dụng design system mới (dark mode, brand gradient)
- [ ] Thêm loading state và toast notifications
- [ ] Đảm bảo responsive trên mobile

---

## PHẦN VI — DESIGN SYSTEM

### DS-1: Nguyên tắc thiết kế

| Thuộc tính | Quyết định |
|:---|:---|
| **Theme** | Dark mode ưu tiên, có toggle Light mode |
| **Primary Color** | Indigo-Violet gradient (`#6366f1` → `#8b5cf6`) |
| **Accent Color** | Emerald (`#10b981`) cho success, Amber (`#f59e0b`) cho warning |
| **Font** | `Inter` từ Google Fonts |
| **Border Radius** | `0.75rem` (12px) cho cards, `0.5rem` cho inputs |
| **Shadow** | Glassmorphism với `backdrop-blur` và `bg-opacity` |
| **Animation** | `transition-all duration-200` chuẩn, `framer-motion` nếu cần |

### DS-2: Component Atoms cần tạo trước

| Component | Mô tả |
|:---|:---|
| `Button` | variants: primary, secondary, ghost, danger + loading state |
| `Input` | error state, icon left/right, disabled |
| `Badge` | variants: success, warning, error, info, neutral |
| `Modal` | Accessible dialog với backdrop |
| `Spinner` | Loading indicator |
| `Toast` | Notification (dùng react-hot-toast) |
| `Skeleton` | Loading placeholder |
| `DataTable` | Sort, filter, pagination server-side |
| `PageHeader` | Title + breadcrumb + actions |
| `EmptyState` | Minh họa khi không có dữ liệu |

---

## PHẦN VII — LỘ TRÌNH THỰC HIỆN

```
Sprint 1 (Tuần 1): Hạ tầng + Auth
  INF-1: Cấu trúc thư mục
  INF-2: API Client (Axios + interceptors)
  INF-3: Auth Context + Protected Route
  PAGE-1: Login Page (kết nối thật)
  DS-2: Các component atoms cơ bản

Sprint 2 (Tuần 2): Admin Core
  PAGE-2: Admin Dashboard
  PAGE-3: Product Management (list + form)
  PAGE-7: Batch Management

Sprint 3 (Tuần 3): Customer Portal
  PAGE-4: Customer Dashboard + Ownership flow
  PAGE-5: QR Scanner + Timeline
  PAGE-8: Warranty (form submit, chờ BE GET API)

Sprint 4 (Tuần 4): AI + Polish
  PAGE-6: AI Hybrid Search
  PAGE-9: Location Map
  PAGE-10: User Management
  PAGE-11: Nâng cấp UI Auth pages
  QA: Cross-browser, responsive, accessibility
```

---

*Tài liệu lập kế hoạch — không bao gồm code. Mọi quyết định triển khai sẽ cập nhật trong sprint notes.*
