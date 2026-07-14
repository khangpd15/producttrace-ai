package dto

type DashboardAlert struct {
	ID          string `json:"id"`
	Type        string `json:"type"`        // "DANGER" (đỏ) or "WARNING" (cam)
	Title       string `json:"title"`       // e.g., "Phát hiện lô sản phẩm hết hạn"
	Description string `json:"description"` // e.g., "Lô hàng BATCH-2026-SP12 (Sơn Spec) đã hết hạn sử dụng. Yêu cầu khóa mã QR."
}
