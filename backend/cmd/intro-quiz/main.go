package main

import (
        "log"
        "net/http"

        "github.com/gin-gonic/gin"
        swaggerFiles "github.com/swaggo/files"
        ginSwagger "github.com/swaggo/gin-swagger"

	"intro-quiz/backend/docs"
        "intro-quiz/backend/internal/config"
        "intro-quiz/backend/internal/handler"
)

// CORSMiddleware sets CORS headers to allow requests from the frontend.
func CORSMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                origin := c.GetHeader("Origin")
                if origin == "http://localhost" || origin == "http://localhost:3000" || origin == "http://localhost:5173" {
                        c.Header("Access-Control-Allow-Origin", origin)
                }
                c.Header("Access-Control-Allow-Credentials", "true")
                c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
                c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

                if c.Request.Method == http.MethodOptions {
                        c.AbortWithStatus(204)
                        return
                }
                c.Next()
        }
}

// @title Intro Quiz API
// @version 1.0
// @description This is the REST API for the Intro Quiz backend.
// @BasePath /
func main() {
        config.LoadEnv()
        router := gin.Default()
        router.Use(CORSMiddleware())

	docs.SwaggerInfo.BasePath = "/"

	router.GET("/ws", handler.WSHandler)
	router.GET("/api/youtube/test", handler.YouTubeTestHandler)
	router.GET("/api/youtube/random", handler.YouTubeRandomHandler)
	router.GET("/api/hello", handler.HelloHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Listening on :8080")
	log.Fatal(router.Run(":8080"))
}
