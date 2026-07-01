package request

import "github.com/google/uuid"

// CreateProductItemRequest là body JSON gửi lên khi tạo một product item mới.
//
// Các trường được tự động sinh tại tầng repository:
//   - item_code:          PTA-{YYMM}-{8 ký tự HEX viết hoa}   vd: PTA-2501-686F493D
//   - serial_number:      SN{14 chữ số ngẫu nhiên}             vd: SN23650263663452
//   - verification_token: chuỗi MD5 hex 32 ký tự thường       vd: fe01edd8ce71cc2911b08118e76e1cd3
type CreateProductItemRequest struct {
	VariantID uuid.UUID `json:"variant_id" binding:"required"`
	BatchID   uuid.UUID `json:"batch_id"`
}
