package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type LocationType string

const (
	TypeWarehouse      LocationType = "WAREHOUSE"
	TypeStore          LocationType = "STORE"
	TypeDealer         LocationType = "DEALER"
	TypeWarrantyCenter LocationType = "WARRANTY_CENTER"
)

// Xử lý JSONB cho GORM
type OpeningHours map[string]struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

func (o OpeningHours) Value() (driver.Value, error) {
	if o == nil { return nil, nil }
	return json.Marshal(o)
}

func (o *OpeningHours) Scan(value interface{}) error {
	if value == nil { return nil }
	b, ok := value.([]byte)
	if !ok { return errors.New("type assertion to []byte failed") }
	return json.Unmarshal(b, &o)
}

type Location struct {
	ID               string       `gorm:"type:uuid;primaryKey;column:id;default:uuid_generate_v4()"`
	OwnerUserID      string       `gorm:"type:uuid;column:owner_user_id"`
	Code             string       `gorm:"type:varchar;unique;column:code"`
	Name             string       `gorm:"type:varchar;not null;column:name"`
	Type             LocationType `gorm:"type:varchar;default:'STORE';column:type"`
	Phone            string       `gorm:"type:varchar;column:phone"`
	Email            string       `gorm:"type:varchar;column:email"`
	Address          string       `gorm:"type:text;column:address"`
	Ward             string       `gorm:"type:varchar;column:ward"`
	District         string       `gorm:"type:varchar;column:district"`
	City             string       `gorm:"type:varchar;column:city"`
	Country          string       `gorm:"type:varchar;default:'Vietnam';column:country"`
	Latitude         float64      `gorm:"type:decimal;column:latitude"`
	Longitude        float64      `gorm:"type:decimal;column:longitude"`
	OpeningHoursJSON OpeningHours `gorm:"type:jsonb;column:opening_hours_json"`
	IsActive         bool         `gorm:"type:boolean;default:true;column:is_active"`
	CreatedAt        time.Time    `gorm:"column:created_at"`
	UpdatedAt        time.Time    `gorm:"column:updated_at"`
	IsDeleted        bool         `gorm:"type:boolean;default:false;column:is_deleted"`
}

func (Location) TableName() string {
	return "locations"
}

// LocationWithDistance được dùng cho kết quả truy vấn FindNearby,
// kết hợp toàn bộ thông tin Location cộng thêm khoảng cách tính bằng mét.
type LocationWithDistance struct {
	Location
	DistanceMeters float64 `gorm:"column:distance_meters"`
}