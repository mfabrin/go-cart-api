package cart

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store CartStore
}

type createCartRequest struct {
	EcommerceID string `json:"ecommerce_id" binding:"required"`
	CustomerID  string `json:"customer_id" binding:"required"`
	Status      Status `json:"status" binding:"required,oneof=created building checkout"`
}

type updateCartRequest struct {
	EcommerceID string              `json:"ecommerce_id" binding:"required"`
	CustomerID  string              `json:"customer_id" binding:"required"`
	ItemList    []updateCartProduct `json:"item_list" binding:"required,min=1,dive"`
}

type updateCartProduct struct {
	ProductSKU   string    `json:"product_sku" binding:"required"`
	ProductName  string    `json:"product_name" binding:"required"`
	FileType     FileType  `json:"file_type" binding:"required,oneof=PDF AI PSD"`
	Quantity     int       `json:"quantity" binding:"required,gt=0"`
	DeliveryDate time.Time `json:"delivery_date" binding:"required"`
}

type checkoutCartRequest struct {
	EcommerceID string `json:"ecommerce_id" binding:"required"`
	CustomerID  string `json:"customer_id" binding:"required"`
}

type cartResponse struct {
	Cart
	ItemList   []productResponse `json:"item_list"`
	TotalPrice float64           `json:"total_price"`
}

type productResponse struct {
	Product
	UnitPrice float64 `json:"unit_price"`
	Price     float64 `json:"price"`
}

func NewHandler(store CartStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) GetCart(c *gin.Context) {
	ecommerceID := c.Param("ecommerce_id")
	customerID := c.Param("customer_id")

	cart, err := h.store.Get(ecommerceID, customerID)

	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	itemList := make([]productResponse, 0, len(cart.ItemList))

	for _, p := range cart.ItemList {
		unitPrice := p.UnitPrice(cart.DateCheckout)
		itemList = append(itemList, productResponse{
			Product:   p,
			UnitPrice: unitPrice,
			Price:     unitPrice * float64(p.Quantity),
		})
	}

	response := cartResponse{
		Cart:       cart,
		ItemList:   itemList,
		TotalPrice: cart.TotalPrice(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateCart(c *gin.Context) {
	var body createCartRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.store.Create(body.EcommerceID, body.CustomerID, body.Status)

	if errors.Is(err, ErrAlreadyExists) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart already exists"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	c.JSON(http.StatusCreated, cart)
}

func (h *Handler) UpdateCart(c *gin.Context) {
	var body updateCartRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	itemList := make([]Product, 0, len(body.ItemList))

	for _, item := range body.ItemList {
		itemList = append(itemList, Product{
			ProductSKU:   item.ProductSKU,
			ProductName:  item.ProductName,
			FileType:     item.FileType,
			Quantity:     item.Quantity,
			DeliveryDate: item.DeliveryDate,
		})
	}

	request := Cart{
		EcommerceID: body.EcommerceID,
		CustomerID:  body.CustomerID,
		Status:      StatusBuilding,
		ItemList:    itemList,
	}

	updatedCart, err := h.store.Update(request)

	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	c.JSON(http.StatusOK, updatedCart)
}

func (h *Handler) CheckoutCart(c *gin.Context) {
	var body checkoutCartRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, getErr := h.store.Get(body.EcommerceID, body.CustomerID)

	if errors.Is(getErr, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if getErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	now := time.Now()

	cart.Status = StatusCheckout
	cart.DateCheckout = &now

	updatedCart, updateErr := h.store.Update(cart)

	if errors.Is(updateErr, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if updateErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	c.JSON(http.StatusOK, updatedCart)
}

func (h *Handler) DeleteCart(c *gin.Context) {
	ecommerceID := c.Param("ecommerce_id")
	customerID := c.Param("customer_id")

	err := h.store.Delete(ecommerceID, customerID)

	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	c.Status(http.StatusNoContent)
}
