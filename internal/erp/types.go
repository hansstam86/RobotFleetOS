package erp

import "time"

type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "draft"
	OrderStatusSubmitted OrderStatus = "submitted"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          string     `json:"id"`
	OrderRef    string     `json:"order_ref"`
	CustomerRef string     `json:"customer_ref,omitempty"`
	SKU         string     `json:"sku"`
	Quantity    int        `json:"quantity"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      OrderStatus `json:"status"`
	MESOrderID  string     `json:"mes_order_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
}
