package dto

import "github.com/google/uuid"

// STEP 1: Request OTP 

// CustomerRequestOTPReq - Customer quét QR: chỉ cần QR code.
// Thông tin chủ sở hữu (name, email, phone) tự lấy từ JWT profile.
type CustomerRequestOTPReq struct {
	QRCode string `json:"qr_code" binding:"required"`
}

// AdminRequestOTPReq - Admin quét QR thay cho khách: cần điền đầy đủ thông tin khách hàng.
type AdminRequestOTPReq struct {
	QRCode     string `json:"qr_code"     binding:"required"`
	OwnerName  string `json:"owner_name"  binding:"required"`
	OwnerEmail string `json:"owner_email" binding:"required,email"`
	OwnerPhone string `json:"owner_phone"`
}

// STEP 2: Verify OTP & Register 

// CustomerVerifyAndRegisterReq - Customer xác minh OTP: chỉ cần OTP và product ID.
// Thông tin chủ sở hữu tự lấy từ profile.
type CustomerVerifyAndRegisterReq struct {
	OTP       string    `json:"otp"        binding:"required"`
	ProductID uuid.UUID `json:"product_id" binding:"required"`
}

// AdminVerifyAndRegisterReq - Admin xác minh OTP thay cho khách: phải kèm đầy đủ thông tin.
type AdminVerifyAndRegisterReq struct {
	OTP        string    `json:"otp"         binding:"required"`
	ProductID  uuid.UUID `json:"product_id"  binding:"required"`
	OwnerName  string    `json:"owner_name"  binding:"required"`
	OwnerEmail string    `json:"owner_email" binding:"required,email"`
	OwnerPhone string    `json:"owner_phone"`
}
