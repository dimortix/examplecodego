package planner

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InteriorDesignRequest struct {
	RoomType string `json:"room_type" binding:"required"`
	Style    string `json:"style" binding:"required"`
}

type InteriorDesignResponse struct {
	URL string `json:"url"`
}

func GenerateInteriorHandler(c *gin.Context) {
	var req InteriorDesignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	log.Printf("Получен запрос на генерацию интерьера: %+v", req)

	aiPlanner := NewAIPlanner()

	url, err := aiPlanner.GenerateInteriorDesign(req.RoomType, req.Style)
	if err != nil {
		log.Printf("Ошибка при генерации интерьера: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать дизайн интерьера"})
		return
	}

	c.JSON(http.StatusOK, InteriorDesignResponse{
		URL: url,
	})
}
