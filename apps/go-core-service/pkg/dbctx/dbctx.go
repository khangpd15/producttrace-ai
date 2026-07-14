// Package dbctx provides a single, shared way to pass a *gorm.DB transaction
// through context.Context across module boundaries.
//
// Trước đây mỗi package repository (product, product_variant,
// product_attribute_value...) tự khai báo type txKey struct{} + InjectTx +
// GetDB riêng. Vì txKey là unexported type của từng package, context value
// set bởi package A không được package B nhận ra (dù cùng tên txKey), khiến
// các câu lệnh SQL "tưởng" đang chạy trong transaction dùng chung lại thực
// ra chạy trên defaultDB (auto-commit riêng lẻ) -> mất tính atomic, dẫn tới
// xoá dở dang (xoá product nhưng variant/attribute_value còn, hoặc ngược lại).
//
// Toàn bộ repository liên quan tới product -> product_variant ->
// attribute_value phải dùng chung package này để đảm bảo cascade
// create/delete nằm trong đúng 1 transaction.
package dbctx

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// InjectTx gắn transaction *gorm.DB vào context để các repository khác nhau
// (dù ở package khác nhau) đều lấy ra được cùng một transaction.
func InjectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetDB trả về transaction nếu context có, ngược lại fallback về defaultDB.
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}
