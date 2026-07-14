# Frontend Batch Module — Backend Integration Plan

## Background

The FE module Batch already has a strong foundation:
- `batch.api.ts` — all 11 backend endpoints are mapped correctly ✅
- `batch.types.ts` — all DTOs correctly map to BE responses ✅
- 4 hooks (`useBatchList`, `useBatchDetail`, `useBatchProducts`, `useBatchHistory`) — all call real APIs ✅
- `BatchListPage.tsx` (1349 lines) — large monolithic page with a **side drawer** for: CREATE, VIEW, EDIT_STATUS, EXPORT, TRACE (DrawerTracePanel), PRODUCTS (DrawerProductsPanel), HISTORY (DrawerHistoryPanel)
- `useTraceSearch` + `trace.api.ts` — already wired, calls `GET /api/trace/search` ✅

---

## Gap Analysis

### ✅ Already working
| Feature | Status |
|---|---|
| Batch list (`GET /batches`) | ✅ Real API, paginated, filtered |
| Batch search (`GET /batches/search`) | ✅ API exists, not used in BatchListPage (uses getList instead) |
| Batch detail drawer (`GET /batches/:batchCode`) | ✅ Real API |
| Create batch (`POST /batches`) | ✅ With product/variant dropdown |
| Update status (`PATCH /batches/:id/status`) | ✅ Real API |
| Export batch (`POST /batches/:id/export`) | ✅ Real API |
| Delete batch (`DELETE /batches/:id`) | ✅ Real API |
| Export QR PDF (`GET /batches/export-qr/:id`) | ✅ Real API, downloads PDF |
| Batch products drawer (paginated) | ✅ Real API with pagination |
| Batch history drawer | ✅ Real API (ADMIN/STAFF only) |
| Trace search in drawer | ✅ Calls `GET /trace/search` with `batch_code` as `code` |
| All status badges, loading, error states | ✅ Implemented |

### ❌ Missing / Not implemented
| Feature | Gap | Impact |
|---|---|---|
| **Batch Events** (`GET /batches/:id/events`) | No hook, no UI | No events in Batch detail |
| **Dedicated Batch Events page** | Not required by user, but drawer should show it | Minor |
| **Dedicated `/batches/:id/products` full page** | Only in drawer currently | User requested separate page |
| **Dedicated `/batches/:id/trace` full page** | Only in drawer currently | User requested separate page |
| **`GET /batches/:id/events` hook** | Missing `useBatchEvents` hook | Cannot show Batch timeline |
| **Trace History full page** | `DrawerTracePanel` calls `trace/search` with batch_code which gives a 400 when batch_code is not a product item code | Logic error — see below |

### ⚠️ Logic Issues Found

1. **`DrawerTracePanel` calls `trace/search` with `batch.batch_code`** — This is a bug. `GET /api/trace/search?code=APL-2026-0001` searches for a **product item** (product_item_id OR serial_number). A batch code is NOT a product item code. This will return no results or 404 for most batches.
   - **Correct approach**: Use `GET /api/batches/:id/events` for batch-level event timeline, OR display all items and let user pick one item to trace.

2. **`DrawerTracePanel` "Trace" for a batch** — According to backend: Trace search supports searching by `item_code` OR `serial_number`. It also works when the product item has a `batch_id`, the API will include batch-level events too. But the input must be an item code/serial, not a batch code.

3. **`useBatchHistory` is called for all roles** but backend allows it only for ADMIN/STAFF. The DrawerHistoryPanel should be role-guarded. Currently it will show a 403 error for DEALER users.

---

## Proposed Changes

### 1. New hook: `useBatchEvents`
Add hook to call `GET /api/batches/:id/events`. Already mapped in `batch.api.ts`.

#### [NEW] `useBatchEvents.ts`
Path: `src/features/batch/hooks/useBatchEvents.ts`
- Returns `events: BatchEventDTO[]`, `isLoading`, `error`, `refetch`

---

### 2. Fix `DrawerTracePanel` — Use `GET /batches/:id/events` instead of `trace/search`

`DrawerTracePanel` currently tries to use `useTraceSearch` with `batch.batch_code` as the search code. This is semantically wrong — the trace search API works with **product item codes**, not batch codes.

**Fix**: Replace `DrawerTracePanel` to use `useBatchEvents(batch.id)` to show the batch-level event timeline. This is the correct, working endpoint.

The "Truy xuất nguồn gốc" (full trace history) for a batch will then be: show the batch events, and let the user navigate to a product item's trace via the Products panel.

#### [MODIFY] `BatchListPage.tsx`
- Fix `DrawerTracePanel` to use `useBatchEvents` instead of `useTraceSearch`
- Add role guard to History button (only show for ADMIN/STAFF)
- Add "View full page" button in Products drawer → navigates to `/batches/:id/products`
- Add "View full page" button in Trace drawer → navigates to `/batches/:id/trace`

---

### 3. New page: Batch Products Page
Full-page view of all product items in a batch.

#### [NEW] `BatchProductsPage.tsx`
Path: `src/admin/pages/batches/BatchProductsPage.tsx`
- URL: `/batches/:batchId/products`
- Uses `useBatchProducts(batchId, params)` — real API
- Full table with: itemCode, serialNumber, status badge, currentLocation, createdAt
- Filter by status, keyword search
- Pagination
- "← Quay lại" button
- Each row has "Xem truy vết" button → links to `/batches/:batchId/trace?itemCode=XXX`

---

### 4. New page: Batch Trace History Page
Full-page view of event timeline for a batch (using `GET /batches/:id/events`).

#### [NEW] `BatchTracePage.tsx`
Path: `src/admin/pages/batches/BatchTracePage.tsx`
- URL: `/batches/:batchId/trace`
- Uses `useBatchEvents(batchId)` — real API
- Shows batch info at top (uses `useBatchDetail` with batch code from state or re-fetch)
- Timeline view of all `BatchEventDTO` events (event_name, detail, created_at)
- Optional: if query param `?itemCode=XXX` is provided, also shows product item timeline via `useTraceSearch`
- "← Quay lại" button

---

### 5. Update AppRouter — Add new routes

#### [MODIFY] `AppRouter.tsx`
Add two new lazy routes under the protected AdminLayout group:
```
/batches/:batchId/products  → BatchProductsPage
/batches/:batchId/trace     → BatchTracePage
```

---

### 6. Add `useBatchEvents` hook

#### [NEW] `useBatchEvents.ts`
```typescript
export function useBatchEvents(batchId: string | null | undefined): {
  events: BatchEventDTO[];
  isLoading: boolean;
  error: string | null;
  refetch: () => void;
}
```

---

## Files Summary

| Action | File | Description |
|---|---|---|
| [NEW] | `src/features/batch/hooks/useBatchEvents.ts` | Hook for `GET /batches/:id/events` |
| [MODIFY] | `src/admin/pages/batches/BatchListPage.tsx` | Fix DrawerTracePanel, add role guard for history, add nav buttons |
| [NEW] | `src/admin/pages/batches/BatchProductsPage.tsx` | Full product items page |
| [NEW] | `src/admin/pages/batches/BatchTracePage.tsx` | Full trace/events timeline page |
| [MODIFY] | `src/routes/AppRouter.tsx` | Add new routes |

---

## Known Limitations (Backend constraints)

> [!IMPORTANT]
> These are existing backend constraints. They will be documented in the UI (not faked).

1. **`GET /batches/:id/events` returns `BatchEventDTO[]` with minimal fields** (`event_name`, `detail`, `created_at`) — no location, no actor. This is what the backend provides.

2. **`GET /batches/:id/history` is ADMIN/STAFF only** — DEALER role gets 403. The history button will be hidden for DEALER.

3. **Trace search (`GET /trace/search?code=XXX`) requires a product item code** (item_code or serial_number) — not a batch code. To view individual item trace history, the user must navigate to the Products page first and then open trace for a specific item.

4. **Export PDF/Excel (trace) are async** — backend may return 202 with a job ID. Full async polling is out of scope; we show the job status returned.

5. **`BatchEventDTO` does not include actor/location** — these are shown as available in the batch events endpoint (`event_name`, `detail`, `created_at` only).

---

## Verification Plan

### Automated
- No unit tests scope (no test infrastructure configured for this module)

### Manual
1. Navigate to `/batches` — list loads with real data, stats show correctly
2. Click "Truy xuất nguồn gốc" (Activity icon) on a batch — opens DrawerTracePanel with batch events from API
3. Click "Danh sách sản phẩm" (Package icon) → opens DrawerProductsPanel with items
4. From Products panel, click "Xem toàn trang" → navigates to `/batches/:id/products`
5. From Products page, each row "Xem truy vết" → navigates to `/batches/:id/trace?itemCode=XXX`
6. `/batches/:id/trace` page shows batch event timeline
7. If `?itemCode` is in URL, also shows product item trace timeline via `useTraceSearch`
8. History button hidden for DEALER role
9. Create/update/delete/export still work as before
