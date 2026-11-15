package cart

import (
	"errors"
	"time"
)

type Status string

const (
	StatusCreated  Status = "created"
	StatusBuilding Status = "building"
	StatusCheckout Status = "checkout"
)

type FileType string

const (
	FileTypePDF FileType = "PDF"
	FileTypePSD FileType = "PSD"
	FileTypeAI  FileType = "AI"
)

type Product struct {
	ProductSKU   string    `json:"product_sku"`
	ProductName  string    `json:"product_name"`
	FileType     FileType  `json:"file_type"`
	Quantity     int       `json:"quantity"`
	DeliveryDate time.Time `json:"delivery_date"`
}

type Cart struct {
	EcommerceID string `json:"ecommerce_id"`
	CustomerID  string `json:"customer_id"`

	Status   Status    `json:"status"`
	ItemList []Product `json:"item_list"`

	DateCheckout *time.Time `json:"date_checkout,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

var (
	ErrNotFound      = errors.New("cart not found")
	ErrAlreadyExists = errors.New("cart already exists")
)

type CartStore interface {
	Create(ecommerceID, customerID string, status Status) (Cart, error)
	Get(ecommerceID, customerID string) (Cart, error)
	Update(cart Cart) (Cart, error)
	Delete(ecommerceID, customerID string) error
}
