package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"intro-quiz/backend/internal/handler"
)

func main() {
	router := gin.Default()
	router.GET("/ws", handler.WSHandler)
	log.Println("Listening on :8080")
	log.Fatal(router.Run(":8080"))
}
