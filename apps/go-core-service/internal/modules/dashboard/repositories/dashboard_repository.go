package repositories

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/dto"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetStats(ctx context.Context) (*dto.DashboardStats, error)
	GetActivities(ctx context.Context, limit int) ([]*dto.DashboardActivity, error)
	GetAlerts(ctx context.Context) ([]*dto.DashboardAlert, error)
	GetProductionSalesChart(ctx context.Context) ([]*dto.DashboardChartItem, error)
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

type queryResult struct {
	ID             string
	EventType      string
	Title          string
	RawDescription string
	CreatedAt      time.Time
	ActorName      *string
	BatchCode      *string
	BatchQuantity  *int
	ProductName    *string
	WarrantyCode   *string
}

func (r *dashboardRepository) GetActivities(ctx context.Context, limit int) ([]*dto.DashboardActivity, error) {
	db := GetDB(ctx, r.db)

	var results []queryResult
	query := `
		SELECT 
			e.id, 
			e.event_type, 
			e.title, 
			e.description AS raw_description, 
			e.created_at,
			u.full_name AS actor_name,
			b.batch_code,
			b.quantity AS batch_quantity,
			p.name AS product_name,
			w.warranty_code
		FROM events e
		LEFT JOIN users u ON e.actor_id = u.id
		LEFT JOIN batches b ON e.batch_id = b.id
		LEFT JOIN product_items pi ON e.product_item_id = pi.id
		LEFT JOIN product_variants pv ON pi.variant_id = pv.id
		LEFT JOIN products p ON pv.product_id = p.id
		LEFT JOIN warranties w ON e.product_item_id = w.product_item_id
		ORDER BY e.created_at DESC
		LIMIT ?;
	`

	if err := db.Raw(query, limit).Scan(&results).Error; err != nil {
		return nil, err
	}

	activities := make([]*dto.DashboardActivity, 0, len(results))
	for _, res := range results {
		actorName := ""
		if res.ActorName != nil {
			actorName = *res.ActorName
		}

		batchCode := ""
		if res.BatchCode != nil {
			batchCode = *res.BatchCode
		}

		batchQuantity := 0
		if res.BatchQuantity != nil {
			batchQuantity = *res.BatchQuantity
		}

		productName := ""
		if res.ProductName != nil {
			productName = *res.ProductName
		}

		warrantyCode := ""
		if res.WarrantyCode != nil {
			warrantyCode = *res.WarrantyCode
		}

		desc := formatActivityDescription(
			res.EventType,
			res.Title,
			res.RawDescription,
			actorName,
			batchCode,
			batchQuantity,
			productName,
			warrantyCode,
		)

		activities = append(activities, &dto.DashboardActivity{
			ID:          res.ID,
			EventType:   res.EventType,
			Title:       res.Title,
			Description: desc,
			CreatedAt:   res.CreatedAt,
		})
	}

	return activities, nil
}

func (r *dashboardRepository) GetAlerts(ctx context.Context) ([]*dto.DashboardAlert, error) {
	db := GetDB(ctx, r.db)
	var alerts []*dto.DashboardAlert

	// 1. Query expired batches
	type expiredBatch struct {
		ID          string
		BatchCode   string
		ProductName string
	}
	var expired []expiredBatch
	expiredQuery := `
		SELECT 
			b.id,
			b.batch_code,
			p.name AS product_name
		FROM batches b
		LEFT JOIN product_variants pv ON b.variant_id = pv.id
		LEFT JOIN products p ON pv.product_id = p.id
		WHERE b.expiry_date < NOW() AND b.status IN ('ACTIVE', 'CREATED', 'IN_STOCK', 'SHIPPED', 'IN_TRANSIT', 'DELIVERED') AND b.is_deleted = false;
	`
	if err := db.Raw(expiredQuery).Scan(&expired).Error; err != nil {
		return nil, err
	}

	for _, item := range expired {
		alerts = append(alerts, &dto.DashboardAlert{
			ID:          item.ID,
			Type:        "DANGER",
			Title:       "Phát hiện lô sản phẩm hết hạn",
			Description: fmt.Sprintf("Lô hàng %s (%s) đã hết hạn sử dụng. Yêu cầu khóa mã QR.", item.BatchCode, item.ProductName),
		})
	}

	// 2. Query low stock levels (< 15)
	type lowStock struct {
		ProductName  string
		LocationName string
		StockCount   int
	}
	var lowStockItems []lowStock
	lowStockQuery := `
		SELECT 
			p.name AS product_name,
			l.name AS location_name,
			COUNT(pi.id) AS stock_count
		FROM product_items pi
		JOIN product_variants pv ON pi.variant_id = pv.id
		JOIN products p ON pv.product_id = p.id
		JOIN locations l ON pi.current_location_id = l.id
		WHERE pi.status = 'IN_STOCK' AND pi.is_deleted = false AND l.is_deleted = false
		GROUP BY p.name, l.name
		HAVING COUNT(pi.id) < 15;
	`
	if err := db.Raw(lowStockQuery).Scan(&lowStockItems).Error; err != nil {
		return nil, err
	}

	for i, item := range lowStockItems {
		alerts = append(alerts, &dto.DashboardAlert{
			ID:          fmt.Sprintf("low-stock-%d", i),
			Type:        "WARNING",
			Title:       "Cảnh báo tồn kho dưới mức an toàn",
			Description: fmt.Sprintf("%s tại %s hiện còn dưới 15 chiếc (Tồn thực tế: %d chiếc).", item.ProductName, item.LocationName, item.StockCount),
		})
	}

	return alerts, nil
}

func (r *dashboardRepository) GetProductionSalesChart(ctx context.Context) ([]*dto.DashboardChartItem, error) {
	db := GetDB(ctx, r.db)

	type queryRow struct {
		TimePeriod time.Time
		Value      float64
	}

	// 1. Get Production Volume (Cách A: batches)
	var productionRows []queryRow
	prodQuery := `
		SELECT 
			DATE_TRUNC('month', manufacture_date) AS time_period,
			SUM(quantity) / 1000.0 AS value
		FROM batches
		WHERE is_deleted = false AND manufacture_date IS NOT NULL
		GROUP BY DATE_TRUNC('month', manufacture_date)
		ORDER BY time_period;
	`
	if err := db.Raw(prodQuery).Scan(&productionRows).Error; err != nil {
		return nil, err
	}

	// 2. Get Sales Volume (Cách B: product_items)
	var salesRows []queryRow
	salesQuery := `
		SELECT 
			DATE_TRUNC('month', sold_at) AS time_period,
			COUNT(*) / 1000.0 AS value
		FROM product_items
		WHERE is_deleted = false AND sold_at IS NOT NULL
		GROUP BY DATE_TRUNC('month', sold_at)
		ORDER BY time_period;
	`
	if err := db.Raw(salesQuery).Scan(&salesRows).Error; err != nil {
		return nil, err
	}

	// 3. Merge results by month
	chartMap := make(map[string]*dto.DashboardChartItem)
	var months []string

	for _, row := range productionRows {
		key := row.TimePeriod.Format("2006-01")
		if _, exists := chartMap[key]; !exists {
			chartMap[key] = &dto.DashboardChartItem{
				TimePeriod: row.TimePeriod,
			}
			months = append(months, key)
		}
		chartMap[key].ProductionVolume = row.Value
	}

	for _, row := range salesRows {
		key := row.TimePeriod.Format("2006-01")
		if _, exists := chartMap[key]; !exists {
			chartMap[key] = &dto.DashboardChartItem{
				TimePeriod: row.TimePeriod,
			}
			months = append(months, key)
		}
		chartMap[key].SalesVolume = row.Value
	}

	// Sort months chronologically
	sort.Strings(months)

	chartData := make([]*dto.DashboardChartItem, 0, len(months))
	for _, m := range months {
		chartData = append(chartData, chartMap[m])
	}

	return chartData, nil
}

func formatActivityDescription(
	eventType string,
	title string,
	rawDescription string,
	actorName string,
	batchCode string,
	batchQuantity int,
	productName string,
	warrantyCode string,
) string {
	switch eventType {
	case "WAREHOUSE_IN", "PRODUCED", "PACKED":
		if batchCode != "" {
			action := "nhập kho"
			if eventType == "PRODUCED" {
				action = "sản xuất"
			} else if eventType == "PACKED" {
				action = "đóng gói"
			}
			qtyStr := ""
			if batchQuantity > 0 {
				qtyStr = fmt.Sprintf(". Số lượng: %d chiếc", batchQuantity)
			}
			actorStr := ""
			if actorName != "" {
				actorStr = fmt.Sprintf(" bởi %s", actorName)
			}
			return fmt.Sprintf("Lô hàng %s được %s thành công%s%s.", batchCode, action, actorStr, qtyStr)
		}
	case "REGISTERED", "SALE":
		if productName != "" {
			action := "Đăng ký quyền sở hữu mới"
			if eventType == "SALE" {
				action = "Bán hàng"
			}
			actorStr := ""
			if actorName != "" {
				actorStr = fmt.Sprintf(" bởi khách hàng %s", actorName)
			}
			return fmt.Sprintf("%s thành công cho thiết bị %s%s.", action, productName, actorStr)
		}
	case "WARRANTY_ACTIVE":
		if warrantyCode != "" && productName != "" {
			actorStr := ""
			if actorName != "" {
				actorStr = fmt.Sprintf(" (%s)", actorName)
			}
			return fmt.Sprintf("Kích hoạt bảo hành điện tử chính hãng %s cho %s%s.", warrantyCode, productName, actorStr)
		}
	}

	if rawDescription != "" {
		return rawDescription
	}
	return title
}
