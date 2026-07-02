package response

import (
	"time"

	"github.com/google/uuid"
)

// ProductItemCreateResponse là DTO trả về sau khi tạo product item thành công.
// Trả về đầy đủ 3 trường định danh được sinh tự động:
//
//   - ItemCode:          PTA-{YYMM}-{8 ký tự HEX viết hoa}   vd: PTA-2501-686F493D
//   - SerialNumber:      SN{14 chữ số ngẫu nhiên}             vd: SN23650263663452
//   - VerificationToken: chuỗi MD5 hex 32 ký tự thường       vd: fe01edd8ce71cc2911b08118e76e1cd3
type ProductItemCreateResponse struct {
	ID                uuid.UUID  `json:"id"`
	VariantID         uuid.UUID  `json:"variant_id"`
	BatchID           *uuid.UUID `json:"batch_id,omitempty"`
	ItemCode          string     `json:"item_code"`
	SerialNumber      string     `json:"serial_number"`
	VerificationToken string     `json:"verification_token"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
}
