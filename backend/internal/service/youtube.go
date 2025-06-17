package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
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
	NextPageToken string `json:"nextPageToken"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []struct {
		Snippet struct {
			Title      string `json:"title"`
			ResourceID struct {
				VideoID string `json:"videoId"`
			} `json:"resourceId"`
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

// GetRandomVideoID returns a random video's ID from the specified playlist.
func (s *YouTubeService) GetRandomVideoID(playlistID string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	maxResults := 50

	// Initial request to obtain total results and first page of items
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=%d&playlistId=%s&key=%s", maxResults, playlistID, s.APIKey)
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

	total := data.PageInfo.TotalResults
	if total == 0 || len(data.Items) == 0 {
		return "", fmt.Errorf("no items found")
	}

	// Select a random index across total results
	target := rand.Intn(total)
	page := target / maxResults
	index := target % maxResults

	// If target within first page, use already fetched data
	items := data.Items
	token := data.NextPageToken
	for i := 0; i < page; i++ {
		if token == "" {
			return "", fmt.Errorf("page token missing")
		}
		pageURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=%d&pageToken=%s&playlistId=%s&key=%s", maxResults, token, playlistID, s.APIKey)
		resp, err := http.Get(pageURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("youtube api status: %s", resp.Status)
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return "", err
		}
		items = data.Items
		token = data.NextPageToken
	}

	if index >= len(items) {
		return "", fmt.Errorf("index out of range")
	}

	return items[index].Snippet.ResourceID.VideoID, nil
}
