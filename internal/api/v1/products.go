package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"devarminas/kcal-track-api/internal/domain/product"
)

type productHandler struct {
	productRepository product.Repository
}

func NewProductHandler(repository product.Repository) *productHandler {
	return &productHandler{
		productRepository: repository,
	}
}

const productSearchLimit = 25

type ProductSearchRequest struct {
	Query string `form:"q" binding:"required"`
}

type GetProductByBarcodeRequest struct {
	Barcode string `uri:"barcode" binding:"required"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Brand       string  `json:"brand" binding:"required"`
	Kcal        float64 `json:"kcal" binding:"required"`
	Carbs       float64 `json:"carbs" binding:"required"`
	Protein     float64 `json:"protein" binding:"required"`
	Fat         float64 `json:"fat" binding:"required"`
	Barcode     string  `json:"barcode"`
	ServingSize float64 `json:"serving_size"`
	ServingUnit string  `json:"serving_unit"`
}

func (h *productHandler) SearchProducts(c *gin.Context) {
	var request ProductSearchRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	products, err := h.productRepository.Search(c.Request.Context(), request.Query, productSearchLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *productHandler) GetProductByBarcode(c *gin.Context) {
	var request GetProductByBarcodeRequest
	if err := c.ShouldBindUri(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prod, err := h.productRepository.FindByBarcode(c.Request.Context(), request.Barcode)
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prod)
}

func (h *productHandler) CreateProduct(c *gin.Context) {
	var request CreateProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nutrition := product.Nutrition{
		Calories:      request.Kcal,
		Protein:       request.Protein,
		Fat:           request.Fat,
		Carbohydrates: request.Carbs,
	}

	opts := []product.ProductOption{
		product.WithBrand(request.Brand),
		product.WithBarcode(request.Barcode),
	}

	prod, err := product.NewProduct(request.Name, nutrition, opts...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdProduct, err := h.productRepository.Save(c.Request.Context(), prod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdProduct)
}
