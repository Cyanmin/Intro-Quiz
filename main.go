package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type playlistItemsResponse struct {
	Items []struct {
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY not set")
	}

	playlistID := "PLBCF2DAC6FFB574DE"

	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=5&playlistId=%s&key=%s",
		playlistID,
		apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to request API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API request failed with status: %s", resp.Status)
	}

	var data playlistItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	for i, item := range data.Items {
		fmt.Printf("%d: %s\n", i+1, item.Snippet.Title)
	}
}
