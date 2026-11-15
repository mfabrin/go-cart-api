package httpapi

import (
	"errors"
	"net/http"
	"time"

	"cart-api/internal/cart"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store cart.CartStore
}

type createCartRequest struct {
	EcommerceID string      `json:"ecommerce_id" binding:"required"`
	CustomerID  string      `json:"customer_id" binding:"required"`
	Status      cart.Status `json:"status" binding:"required,oneof=created building checkout"`
}

type updateCartRequest struct {
	EcommerceID string              `json:"ecommerce_id" binding:"required"`
	CustomerID  string              `json:"customer_id" binding:"required"`
	ItemList    []updateCartProduct `json:"item_list" binding:"required,min=1,dive"`
}

type updateCartProduct struct {
	ProductSKU   string        `json:"product_sku" binding:"required"`
	ProductName  string        `json:"product_name" binding:"required"`
	FileType     cart.FileType `json:"file_type" binding:"required,oneof=PDF AI PSD"`
	Quantity     int           `json:"quantity" binding:"required,gt=0"`
	DeliveryDate time.Time     `json:"delivery_date" binding:"required"`
}

type checkoutCartRequest struct {
	EcommerceID string `json:"ecommerce_id" binding:"required"`
	CustomerID  string `json:"customer_id" binding:"required"`
}

type cartResponse struct {
	cart.Cart
	ItemList   []productResponse `json:"item_list"`
	TotalPrice float64           `json:"total_price"`
}

type productResponse struct {
	cart.Product
	UnitPrice float64 `json:"unit_price"`
	Price     float64 `json:"price"`
}

func New(store cart.CartStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) GetCart(c *gin.Context) {
	ecommerceID := c.Param("ecommerce_id")
	customerID := c.Param("customer_id")

	crt, err := h.store.Get(ecommerceID, customerID)

	if errors.Is(err, cart.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	itemList := make([]productResponse, 0, len(crt.ItemList))

	for _, p := range crt.ItemList {
		unitPrice := p.UnitPrice(crt.DateCheckout)
		itemList = append(itemList, productResponse{
			Product:   p,
			UnitPrice: unitPrice,
			Price:     unitPrice * float64(p.Quantity),
		})
	}

	response := cartResponse{
		Cart:       crt,
		ItemList:   itemList,
		TotalPrice: crt.TotalPrice(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateCart(c *gin.Context) {
	var body createCartRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	crt, err := h.store.Create(body.EcommerceID, body.CustomerID, body.Status)

	if errors.Is(err, cart.ErrAlreadyExists) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart already exists"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	c.JSON(http.StatusCreated, crt)
}

func (h *Handler) UpdateCart(c *gin.Context) {
	var body updateCartRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	itemList := make([]cart.Product, 0, len(body.ItemList))

	for _, item := range body.ItemList {
		itemList = append(itemList, cart.Product{
			ProductSKU:   item.ProductSKU,
			ProductName:  item.ProductName,
			FileType:     item.FileType,
			Quantity:     item.Quantity,
			DeliveryDate: item.DeliveryDate,
		})
	}

	request := cart.Cart{
		EcommerceID: body.EcommerceID,
		CustomerID:  body.CustomerID,
		Status:      cart.StatusBuilding,
		ItemList:    itemList,
	}

	updatedCart, err := h.store.Update(request)

	if errors.Is(err, cart.ErrNotFound) {
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

	crt, getErr := h.store.Get(body.EcommerceID, body.CustomerID)

	if errors.Is(getErr, cart.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if getErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	now := time.Now()

	crt.Status = cart.StatusCheckout
	crt.DateCheckout = &now

	updatedCart, updateErr := h.store.Update(crt)

	if errors.Is(updateErr, cart.ErrNotFound) {
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

	if errors.Is(err, cart.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error occurred"})
		return
	}

	c.Status(http.StatusNoContent)
}
