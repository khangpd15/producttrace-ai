package domain

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
type OpeningHour struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type OpeningHours map[string]OpeningHour

func (o OpeningHours) Value() (driver.Value, error) {
	if o == nil {
		return nil, nil
	}
	return json.Marshal(o)
}

func (o *OpeningHours) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("unsupported type for OpeningHours.Scan")
	}
	return json.Unmarshal(b, &o)
}

type Location struct {
	ID               string       `gorm:"type:uuid;primaryKey;column:id"`
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
	GeoLocation      *GeoLocation `gorm:"column:geo_location;type:geography"`
}

func (Location) TableName() string {
	return "locations"
}

type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (g GeoLocation) Value() (driver.Value, error) {
	return fmt.Sprintf("SRID=4326;POINT(%f %f)", g.Longitude, g.Latitude), nil
}

func (g *GeoLocation) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into GeoLocation", value)
	}

	if len(data) == 0 {
		return nil
	}

	// Thử hex decode cho bất kỳ byte slice có độ dài chẵn >= 42 (min EWKB Point)
	// 42 = 21 bytes × 2 hex chars (Point không có SRID), 50 = 25 bytes × 2 (Point có SRID)
	if len(data)%2 == 0 && len(data) >= 42 {
		decoded := make([]byte, len(data)/2)
		n, err := hex.Decode(decoded, data)
		if err == nil && n == len(data)/2 {
			data = decoded
		}
	}

	if len(data) < 25 {
		return fmt.Errorf("invalid EWKB length: %d", len(data))
	}

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if data[0] == 0x00 {
		byteOrder = binary.BigEndian
	} else if data[0] != 0x01 {
		return fmt.Errorf("invalid byte order indicator: %d", data[0])
	}

	wkbType := byteOrder.Uint32(data[1:5])
	hasSRID := (wkbType & 0x20000000) != 0
	pureType := wkbType & 0x1FFFFFFF

	if pureType != 1 {
		return fmt.Errorf("unsupported geometry type: %d (expected Point)", pureType)
	}

	offset := 5
	if hasSRID {
		offset += 4
	}

	if len(data) < offset+16 {
		return fmt.Errorf("insufficient bytes for coordinates: %d", len(data))
	}

	xBits := byteOrder.Uint64(data[offset : offset+8])
	yBits := byteOrder.Uint64(data[offset+8 : offset+16])

	g.Longitude = math.Float64frombits(xBits)
	g.Latitude = math.Float64frombits(yBits)

	return nil
}
