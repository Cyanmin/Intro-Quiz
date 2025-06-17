package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"intro-quiz/backend/internal/config"
	"intro-quiz/backend/internal/handler"
)

func main() {
	config.LoadEnv()
	router := gin.Default()
	router.GET("/ws", handler.WSHandler)
	router.GET("/api/youtube/test", handler.YouTubeTestHandler)
	log.Println("Listening on :8080")
	log.Fatal(router.Run(":8080"))
}
