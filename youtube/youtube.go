package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const BaseUrl = "https://www.googleapis.com/youtube/v3/"

type YouTube struct {
	apiKey string
}

type VideoMetadata struct {
	VideoId          string
	VideoTitle       string
	VideoThumbnail   string
	ViewCount        string
	PublishDate      string
	ChannelTitle     string
	ChannelId        string
	ChannelThumbnail string
	SubscriberCount  string
	VideoCount       string
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
		"statistics",
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

	videoId, ok := item["id"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	snippet, ok := item["snippet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	publishDate, ok := snippet["publishedAt"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	t, err := time.Parse(time.RFC3339, publishDate)
	if err == nil {
		publishDate = t.Format("2006-01-02")
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

	statistics, ok := item["statistics"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	viewCount, ok := statistics["viewCount"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	const channelEndpoint = "channels/"
	properties = []string{
		"snippet",
		"statistics",
	}
	url, err = url.Parse(BaseUrl + channelEndpoint)
	if err != nil {
		return nil, err
	}

	q = url.Query()
	q.Set("key", youtube.apiKey)
	q.Set("part", strings.Join(properties, ","))
	q.Set("id", channelId)
	q.Set("maxResults", "1")
	url.RawQuery = q.Encode()

	res2, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	if res2.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Call to YouTube API failed with Status: %s.", res.Status,
		)
	}

	defer res2.Body.Close()

	body = make(map[string]interface{})

	json.NewDecoder(res2.Body).Decode(&body)

	items, ok = body["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("No results found for %s.", idOrUrl)
	}

	item, ok = items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	snippet, ok = item["snippet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	thumbnails, ok = snippet["thumbnails"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	thumbnail, ok = thumbnails["medium"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	channelThumbnailUrl, ok := thumbnail["url"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	statistics, ok = item["statistics"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	subscriberCount, ok := statistics["subscriberCount"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	videoCount, ok := statistics["videoCount"].(string)
	if !ok {
		return nil, fmt.Errorf("Error in YouTube Data API response format.")
	}

	return &VideoMetadata{
		VideoId:          videoId,
		VideoTitle:       title,
		VideoThumbnail:   thumbnailUrl,
		ViewCount:        viewCount,
		PublishDate:      publishDate,
		ChannelId:        channelId,
		ChannelTitle:     channelTitle,
		ChannelThumbnail: channelThumbnailUrl,
		SubscriberCount:  subscriberCount,
		VideoCount:       videoCount,
	}, nil
}
