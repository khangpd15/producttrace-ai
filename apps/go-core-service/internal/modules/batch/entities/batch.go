package entities

import (
	"time"

	"github.com/google/uuid"
	entities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
)

type Batch struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	VariantID        uuid.UUID  `gorm:"type:uuid;not null" json:"variant_id"`
	BatchCode        string     `gorm:"unique;not null" json:"batch_code"`
	ManufactureDate  *time.Time `json:"manufacture_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	ImportedAt       *time.Time `json:"imported_at"`
	ManufacturerName string     `json:"manufacturer_name"`
	SupplierName     string     `json:"supplier_name"`
	OriginCountry    string     `json:"origin_country"`
	ProductionPlace  string     `gorm:"type:text" json:"production_place"`
	Quantity         int        `gorm:"default:0" json:"quantity"`
	Status           string     `gorm:"default:'ACTIVE'" json:"status"`
	CreatedBy        *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	IsDeleted        bool       `gorm:"default:false" json:"is_deleted"`

	// Relationships
	ProductVariant entities.ProductVariant `gorm:"foreignKey:VariantID;references:ID"`
}

func (Batch) TableName() string {
	return "batches"
}
