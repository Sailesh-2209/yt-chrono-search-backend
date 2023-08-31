package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const BaseUrl = "https://www.googleapis.com/youtube/v3/"

type YouTube struct {
	apiKey string
}

type VideoMetadata struct {
	VideoTitle     string
	VideoThumbnail string
	ChannelTitle   string
	ChannelId      string
}

func NewYouTubeService(apiKey string) *YouTube {
	youtubeService := YouTube{
		apiKey: apiKey,
	}

	return &youtubeService
}

func (youtube *YouTube) GetVideoMetadata(idOrUrl string) (*VideoMetadata, error) {
	const endpoint = "videos/"
	properties := []string{
		"snippet",
	}

	id, err := parseVideoId(idOrUrl)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(BaseUrl + endpoint)
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("key", youtube.apiKey)
	q.Set("part", strings.Join(properties, ","))
	q.Set("id", id)
	q.Set("maxResults", "1")
	url.RawQuery = q.Encode()

	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Call to YouTube API failed with Status: %s.", res.Status,
		)
	}

	defer res.Body.Close()

	body := make(map[string]interface{})

	json.NewDecoder(res.Body).Decode(&body)

	items, ok := body["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("No results found for %s.", idOrUrl)
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	snippet, ok := item["snippet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	title, ok := snippet["title"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	thumbnails, ok := snippet["thumbnails"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	thumbnail, ok := thumbnails["standard"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	thumbnailUrl, ok := thumbnail["url"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	channelId, ok := snippet["channelId"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	channelTitle, ok := snippet["channelTitle"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	return &VideoMetadata{
		VideoTitle:     title,
		VideoThumbnail: thumbnailUrl,
		ChannelId:      channelId,
		ChannelTitle:   channelTitle,
	}, nil
}
