package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// YouTubeService provides methods to interact with YouTube Data API.
type YouTubeService struct {
	APIKey string
}

// NewYouTubeService creates a new YouTubeService.
func NewYouTubeService(key string) *YouTubeService {
	return &YouTubeService{APIKey: key}
}

// playlistItemsResponse represents a subset of the YouTube API response.
type playlistItemsResponse struct {
	Items []struct {
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

// GetFirstVideoTitle returns the first video's title from the given playlist.
func (s *YouTubeService) GetFirstVideoTitle(playlistID string) (string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=1&playlistId=%s&key=%s", playlistID, s.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube api status: %s", resp.Status)
	}
	var data playlistItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Items) == 0 {
		return "", fmt.Errorf("no items found")
	}
	return data.Items[0].Snippet.Title, nil
}
