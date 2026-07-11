package repositories

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/dto"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetStats(ctx context.Context) (*dto.DashboardStats, error)
}

type txKey struct{}

func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetStats(ctx context.Context) (*dto.DashboardStats, error) {
	db := GetDB(ctx, r.db)

	var totalProducts int64
	if err := db.Table("products").Where("is_deleted = false").Count(&totalProducts).Error; err != nil {
		return nil, err
	}

	var totalBatches int64
	if err := db.Table("batches").Where("is_deleted = false").Count(&totalBatches).Error; err != nil {
		return nil, err
	}

	var totalOwnerships int64
	if err := db.Table("ownerships").Count(&totalOwnerships).Error; err != nil {
		return nil, err
	}

	var totalUnderWarranty int64
	if err := db.Table("warranties").Where("status = ?", "ACTIVE").Count(&totalUnderWarranty).Error; err != nil {
		return nil, err
	}

	var totalPendingApproval int64
	if err := db.Table("warranties").Where("status = ?", "CLAIMED").Count(&totalPendingApproval).Error; err != nil {
		return nil, err
	}

	var totalLocations int64
	if err := db.Table("locations").Where("is_deleted = false").Count(&totalLocations).Error; err != nil {
		return nil, err
	}

	return &dto.DashboardStats{
		TotalProducts:        totalProducts,
		TotalBatches:         totalBatches,
		TotalOwnerships:      totalOwnerships,
		TotalUnderWarranty:   totalUnderWarranty,
		TotalPendingApproval: totalPendingApproval,
		TotalLocations:       totalLocations,
	}, nil
}
