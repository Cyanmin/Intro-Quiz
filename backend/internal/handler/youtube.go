package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"intro-quiz/backend/internal/service"
)

// YouTubeTestHandler returns the first video title of a fixed playlist.
// @Summary      Get first video title
// @Description  Retrieve the first video's title from a fixed YouTube playlist.
// @Tags         youtube
// @Produce      json
// @Success      200 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/youtube/test [get]
func YouTubeTestHandler(c *gin.Context) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	svc := service.NewYouTubeService(apiKey)
	title, err := svc.GetFirstVideoTitle("PLBCF2DAC6FFB574DE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "title": title})
}
