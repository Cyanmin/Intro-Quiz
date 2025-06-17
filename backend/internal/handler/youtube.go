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

// YouTubeRandomHandler returns a random video ID from the specified playlist.
// @Summary      Get random video ID
// @Description  Retrieve a random video's ID from a YouTube playlist.
// @Tags         youtube
// @Produce      json
// @Param        playlistId  query     string  true  "Playlist ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/youtube/random [get]
func YouTubeRandomHandler(c *gin.Context) {
	playlistID := c.Query("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "playlistId required"})
		return
	}
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	svc := service.NewYouTubeService(apiKey)
	videoID, err := svc.GetRandomVideoID(playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "videoId": videoID})
}
