package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
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
	Items         []struct {
		Snippet struct {
			Title      string `json:"title"`
			ResourceID struct {
				VideoID string `json:"videoId"`
			} `json:"resourceId"`
		} `json:"snippet"`
	} `json:"items"`
}

// VideoItem represents a single video ID and title pair.
type VideoItem struct {
	ID    string
	Title string
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

// GetFirstVideoID returns the first video's ID from the given playlist.
func (s *YouTubeService) GetFirstVideoID(playlistID string) (string, error) {
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
	return data.Items[0].Snippet.ResourceID.VideoID, nil
}

// ListPlaylistVideos retrieves all video IDs and titles from the playlist.
func (s *YouTubeService) ListPlaylistVideos(playlistID string) ([]VideoItem, error) {
	var videos []VideoItem
	pageToken := ""
	for {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=50&playlistId=%s&key=%s&pageToken=%s", playlistID, s.APIKey, pageToken)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("youtube api status: %s", resp.Status)
		}
		var data playlistItemsResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, err
		}
		for _, it := range data.Items {
			videos = append(videos, VideoItem{ID: it.Snippet.ResourceID.VideoID, Title: it.Snippet.Title})
		}
		if data.NextPageToken == "" {
			break
		}
		pageToken = data.NextPageToken
	}
	return videos, nil
}

// GetRandomVideo returns a random video's ID and title from the given playlist.
func (s *YouTubeService) GetRandomVideo(playlistID string) (string, string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=50&playlistId=%s&key=%s", playlistID, s.APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("youtube api status: %s", resp.Status)
	}
	var data playlistItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}
	if len(data.Items) == 0 {
		return "", "", fmt.Errorf("no items found")
	}
	// Go 1.20以降はrand.Seedでの初期化は不要です
	indices := rand.Perm(len(data.Items))
	for _, idx := range indices {
		item := data.Items[idx].Snippet
		embeddable, err := CheckEmbeddable(item.ResourceID.VideoID)
		if err != nil {
			continue
		}
		if embeddable {
			return item.ResourceID.VideoID, item.Title, nil
		}
	}
	return "", "", fmt.Errorf("no embeddable videos found")
}

// CheckEmbeddable verifies whether the specified video can be embedded.
func CheckEmbeddable(videoID string) (bool, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return false, fmt.Errorf("YOUTUBE_API_KEY not set")
	}
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=status&id=%s&key=%s", videoID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("youtube api status: %s", resp.Status)
	}
	var result struct {
		Items []struct {
			Status struct {
				Embeddable bool `json:"embeddable"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	if len(result.Items) == 0 {
		return false, fmt.Errorf("no items found")
	}
	return result.Items[0].Status.Embeddable, nil
}
