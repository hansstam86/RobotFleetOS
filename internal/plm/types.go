package plm

import "time"

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusReleased ProductStatus = "released"
	ProductStatusObsolete ProductStatus = "obsolete"
)

type Product struct {
	ID        string        `json:"id"`
	SKU       string        `json:"sku"`
	Name      string        `json:"name"`
	Revision  string        `json:"revision"`
	Status    ProductStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type BOMLine struct {
	ID              string    `json:"id"`
	ParentProductID string    `json:"parent_product_id"`
	ChildSKU        string    `json:"child_sku"`
	ChildRevision   string    `json:"child_revision,omitempty"`
	Quantity        float64   `json:"quantity"`
	LineNumber      int       `json:"line_number,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type ECOStatus string

const (
	ECOStatusDraft       ECOStatus = "draft"
	ECOStatusApproved    ECOStatus = "approved"
	ECOStatusImplemented ECOStatus = "implemented"
)

type ECO struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ProductID   string    `json:"product_id,omitempty"`
	Status      ECOStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
