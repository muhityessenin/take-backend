package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"warehouse-backend/internal/model"
	"warehouse-backend/internal/repo"
)

type ItemHandler struct {
	Repo *repo.ItemRepository
}

func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{
		Repo: repo.NewItemRepository(db),
	}
}

func (h *ItemHandler) AddItem(c *gin.Context) {
	var item model.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.AddItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ItemHandler) GetItems(c *gin.Context) {
	brand := c.Query("brand")

	var (
		items []model.Item
		err   error
	)

	if brand != "" {
		items, err = h.Repo.GetItemsByBrand(brand)
	} else {
		items, err = h.Repo.GetAllItems()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список товаров"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *ItemHandler) MakeSale(c *gin.Context) {
	var req struct {
		ItemID   uint   `json:"itemId"`
		Quantity int    `json:"quantity"`
		Customer string `json:"customer"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	sale, err := h.Repo.MakeSale(req.ItemID, req.Quantity, req.Customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sale)
}
func (h *ItemHandler) GetTodaySales(c *gin.Context) {
	sales, err := h.Repo.GetTodaySales()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить продажи за сегодня"})
		return
	}
	c.JSON(http.StatusOK, sales)
}

func (h *ItemHandler) GetTop5BestSellers(c *gin.Context) {
	top, err := h.Repo.GetTop5BestSellers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить топ продаж"})
		return
	}
	c.JSON(http.StatusOK, top)
}

func (h *ItemHandler) GetSales(c *gin.Context) {
	brand := c.Query("brand")

	var (
		sales []model.Sale
		err   error
	)

	if brand != "" {
		sales, err = h.Repo.GetSalesByBrand(brand)
	} else {
		sales, err = h.Repo.GetTodaySales() // или можешь сделать GetAllSales()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить продажи"})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (h *ItemHandler) UpdateItem(c *gin.Context) {
	idParam := c.Param("id")
	var updates map[string]interface{}

	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	updatedItem, err := h.Repo.UpdateItem(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении товара"})
		return
	}

	c.JSON(http.StatusOK, updatedItem)
}
