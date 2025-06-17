package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"intro-quiz/backend/docs"
	"intro-quiz/backend/internal/config"
	"intro-quiz/backend/internal/handler"
)

// @title Intro Quiz API
// @version 1.0
// @description This is the REST API for the Intro Quiz backend.
// @BasePath /
func main() {
	config.LoadEnv()
	router := gin.Default()

	docs.SwaggerInfo.BasePath = "/"

	router.GET("/ws", handler.WSHandler)
	router.GET("/api/youtube/test", handler.YouTubeTestHandler)
	router.GET("/api/youtube/random", handler.YouTubeRandomHandler)
	router.GET("/api/hello", handler.HelloHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Listening on :8080")
	log.Fatal(router.Run(":8080"))
}
