package dtos

// AdminStatsResponse contains platform statistics.
type AdminStatsResponse struct {
	TotalUsers      int64  `json:"total_users"`
	TotalSellers    int64  `json:"total_sellers"`
	TotalProducts   int64  `json:"total_products"`
	TotalOrders     int64  `json:"total_orders"`
	PendingOrders   int64  `json:"pending_orders"`
	PaidOrders      int64  `json:"paid_orders"`
	ShippedOrders   int64  `json:"shipped_orders"`
	DeliveredOrders int64  `json:"delivered_orders"`
	CancelledOrders int64  `json:"cancelled_orders"`
	TotalRevenue    string `json:"total_revenue"`
}
