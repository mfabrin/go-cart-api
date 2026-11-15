package cart

import (
	"errors"
	"net/http"

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

	c.JSON(http.StatusOK, cart)
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
