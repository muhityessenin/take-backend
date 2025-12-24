package handler

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"warehouse-backend/internal/model"
	"warehouse-backend/internal/repo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxImageSize = 30 << 20 // 30 MB
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,

	// iPhone
	".heic": true,
	".heif": true,
}

type ItemHandler struct {
	Repo *repo.ItemRepository
}

func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{
		Repo: repo.NewItemRepository(db),
	}
}
func validateImage(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > maxImageSize {
		return fmt.Errorf("file too large (max 30MB)")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	return nil
}

func saveImageLocally(
	file multipart.File,
	filename string,
) (string, error) {

	// создаём директорию если нет
	dir := "uploads/items"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// безопасное имя файла
	ext := filepath.Ext(filename)
	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	fullPath := filepath.Join(dir, newName)

	out, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	// то, что сохраняем в БД
	return "/" + fullPath, nil
}
func (h *ItemHandler) AddItem(c *gin.Context) {
	name := c.PostForm("name")
	partNumber := c.PostForm("partNumber")
	brand := c.PostForm("brand")
	modelName := c.PostForm("model")
	stock, _ := strconv.Atoi(c.PostForm("stock"))
	price, _ := strconv.Atoi(c.PostForm("price"))
	wholesalePrice, _ := strconv.Atoi(c.PostForm("wholesalePrice"))

	item := model.Item{
		Name:           name,
		PartNumber:     partNumber,
		Brand:          brand,
		Model:          modelName,
		Stock:          stock,
		Price:          price,
		WholesalePrice: wholesalePrice,
	}

	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]

		for _, fileHeader := range files {

			if err := validateImage(fileHeader); err != nil {
				log.Println("Image skipped:", err)
				continue
			}

			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			url, err := saveImageLocally(file, fileHeader.Filename)
			file.Close()

			if err != nil {
				log.Println("Save error:", err)
				continue
			}

			item.Images = append(item.Images, model.ItemImage{
				URL: url,
			})
		}
	}

	if err := h.Repo.AddItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add item"})
		return
	}

	c.JSON(http.StatusOK, item)
}
func deleteLocalImage(path string) {
	if path == "" {
		return
	}

	// path в БД вида /uploads/items/xxx.jpg
	fullPath := "." + path
	if err := os.Remove(fullPath); err != nil {
		log.Println("Failed to delete image:", fullPath, err)
	}
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
		sales, err = h.Repo.GetAllSales() // или можешь сделать GetAllSales()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить продажи"})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (h *ItemHandler) UpdateItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Ручной разбор формы, потому что multipart/form-data
	updates := make(map[string]interface{})

	if name := c.PostForm("name"); name != "" {
		updates["name"] = name
	}
	if partNumber := c.PostForm("partNumber"); partNumber != "" {
		updates["part_number"] = partNumber
	}
	if brand := c.PostForm("brand"); brand != "" {
		updates["brand"] = brand
	}
	if modelName := c.PostForm("model"); modelName != "" {
		updates["model"] = modelName
	}
	if stockStr := c.PostForm("stock"); stockStr != "" {
		if stock, err := strconv.Atoi(stockStr); err == nil {
			updates["stock"] = stock
		}
	}
	if priceStr := c.PostForm("price"); priceStr != "" {
		if price, err := strconv.Atoi(priceStr); err == nil {
			updates["price"] = price
		}
	}
	if wholesaleStr := c.PostForm("wholesalePrice"); wholesaleStr != "" {
		if wholesale, err := strconv.Atoi(wholesaleStr); err == nil {
			updates["wholesale_price"] = wholesale
		}
	}

	// Обработка изображений
	form, _ := c.MultipartForm()
	files := form.File["images"]
	var images []model.ItemImage

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		url, err := saveImageLocally(file, fileHeader.Filename)
		file.Close()

		if err == nil {
			images = append(images, model.ItemImage{URL: url})
		}
	}

	if len(images) > 0 {
		updates["images"] = images
	}

	// Обновление в репозитории
	updatedItem, err := h.Repo.UpdateItem(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedItem)
}
