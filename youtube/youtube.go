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
	ChannelCustomUrl string
	SubscriberCount  string
	VideoCount       string
}

type VideoList struct {
	Count  int
	Videos []*VideoMetadata
}

type ChannelPlaylist struct {
	Id        string
	Title     string
	Thumbnail string
}

type ChannelPlaylists struct {
	Count     int
	Playlists []*ChannelPlaylist
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
			"call to YouTube API %s endpoint failed with message %s",
			endpoint,
			res.Status,
		)
	}

	defer res.Body.Close()

	body := make(map[string]interface{})

	json.NewDecoder(res.Body).Decode(&body)

	items, ok := body["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'items' not found",
			url.String(),
		)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no results found for %s", idOrUrl)
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s", url.String(),
		)
	}

	videoId, ok := item["id"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'id' not found in items[0]",
			url.String(),
		)
	}

	snippet, ok := item["snippet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'snippet' not found in items[0]",
			url.String(),
		)
	}

	publishDate, ok := snippet["publishedAt"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'publishedAt' not found in items[0]['snippet']",
			url.String(),
		)
	}

	t, err := time.Parse(time.RFC3339, publishDate)
	if err == nil {
		publishDate = t.Format("2006-01-02")
	}

	title, ok := snippet["title"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'title' not found in items[0]['snippet']",
			url.String(),
		)
	}

	thumbnails, ok := snippet["thumbnails"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'thumbnails' not found in items[0]['snippet']",
			url.String(),
		)
	}

	thumbnail, ok := thumbnails["standard"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'standard' not found in items[0]['snippet']['thumbnails'e",
			url.String(),
		)
	}

	thumbnailUrl, ok := thumbnail["url"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'url' not found in items[0]['snippet']['thumbnails']['standard']",
			url.String(),
		)
	}

	channelId, ok := snippet["channelId"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'channelId' not found in items[0]['snippet']",
			url.String(),
		)
	}

	channelTitle, ok := snippet["channelTitle"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response %s. Key 'channelTitle' not found in items[0]['snippet']",
			url.String(),
		)
	}

	statistics, ok := item["statistics"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'statistics' not found in items[0]",
			url.String(),
		)
	}

	viewCount, ok := statistics["viewCount"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'viewCount' not found in items[0]['statistics']",
			url.String(),
		)
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
			"call to YouTube API %s endpoint failed with Status: %s", channelEndpoint, res.Status,
		)
	}

	body = make(map[string]interface{})

	json.NewDecoder(res2.Body).Decode(&body)

	res2.Body.Close()

	items, ok = body["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'items' not found",
			url.String(),
		)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no channels found for %s", channelId)
	}

	item, ok = items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Unable to index 'items' list",
			url.String(),
		)
	}

	snippet, ok = item["snippet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'snippet' not found in items[0]",
			url.String(),
		)
	}

	customUrl, ok := snippet["customUrl"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'customUrl not found in items[0]['snippet']",
			url.String(),
		)
	}

	thumbnails, ok = snippet["thumbnails"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'thumbnails' not found in items[0]['snippet']",
			url.String(),
		)
	}

	thumbnail, ok = thumbnails["medium"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'medium' not found in items[0]['snippet']['thumbnails']",
			url.String(),
		)
	}

	channelThumbnailUrl, ok := thumbnail["url"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'url' not found in items[0]['snippet']['thumbnails']['medium']",
			url.String(),
		)
	}

	statistics, ok = item["statistics"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'statistics' not found in items[0]",
			url.String(),
		)
	}

	subscriberCount, ok := statistics["subscriberCount"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'subscriberCount' not found in items[0]['statistics']",
			url.String(),
		)
	}

	videoCount, ok := statistics["videoCount"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'videoCount' not found in items[0]['statistics']",
			url.String(),
		)
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
		ChannelCustomUrl: customUrl,
		SubscriberCount:  subscriberCount,
		VideoCount:       videoCount,
	}, nil
}

func (youtube *YouTube) GetChannelVideos(
	channelId string, videoId string, pageToken string, loadPrev bool,
) (*VideoList, error) {
	const endpoint = "search/"

	requestUrl, err := url.Parse(BaseUrl + endpoint)
	if err != nil {
		return nil, err
	}

	q := requestUrl.Query()
	q.Set("key", youtube.apiKey)
	q.Set("channelId", channelId)
	q.Set("maxResults", "10")
	q.Set("type", "video")
	q.Set("order", "date")
	if pageToken == "" {
		pageToken, err = findPageToken(videoId, channelId, youtube.apiKey)
		if err != nil {
			return nil, err
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
	}
	requestUrl.RawQuery = q.Encode()

	res, err := http.Get(requestUrl.String())
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf(
			"call to YouTube API endpoint %s failed with message %s",
			endpoint,
			res.Status,
		)
	}

	body := make(map[string]interface{})

	json.NewDecoder(res.Body).Decode(&body)

	res.Body.Close()

	result := []*VideoMetadata{}

	if loadPrev {
		items, ok := body["items"].([]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"error in YouTube Data API response from %s. Key 'items' not found",
				requestUrl.String(),
			)
		}
		if len(items) == 0 {
			return nil, fmt.Errorf(
				"no videos found for channel with channel id %s",
				channelId,
			)
		}

		appendFlag := false

		result := []*VideoMetadata{}

		for i := 0; i < len(items); i++ {
			item, ok := items[i].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]",
					requestUrl.String(),
					i,
				)
			}

			id, ok := item["id"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['id']",
					requestUrl.String(),
					i,
				)
			}

			video, ok := id["videoId"].(string)
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['videoId']",
					requestUrl.String(),
					i,
				)
			}

			if appendFlag {
				videoMetadata, err := youtube.GetVideoMetadata(video)
				if err != nil {
					return nil, fmt.Errorf(
						"error in finding metadata for video %s",
						video,
					)
				}
				result = append(result, videoMetadata)
			}

			if video == videoId {
				appendFlag = true
			}
		}

		nextPageToken, nextPageExists := body["nextPageToken"].(string)
		if !nextPageExists {
			return &VideoList{
				Count:  len(result),
				Videos: result,
			}, nil
		}

		q := requestUrl.Query()
		q.Set("key", youtube.apiKey)
		q.Set("channelId", channelId)
		q.Set("maxResults", "10")
		q.Set("type", "video")
		q.Set("order", "date")
		q.Set("pageToken", nextPageToken)

		requestUrl.RawQuery = q.Encode()

		res, err = http.Get(requestUrl.String())
		if err != nil {
			return nil, err
		}

		if res.StatusCode != 200 {
			return nil, fmt.Errorf(
				"call to YouTube API endpoint %s failed with message %s",
				endpoint,
				res.Status,
			)
		}

		body = make(map[string]interface{})

		json.NewDecoder(res.Body).Decode(&body)

		res.Body.Close()

		items, ok = body["items"].([]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"error in YouTube Data API response from %s. Key 'items' not found",
				requestUrl.String(),
			)
		}
		if len(items) == 0 {
			return nil, fmt.Errorf(
				"no videos found for channel with channel id %s",
				channelId,
			)
		}

		for i := 0; i < len(items); i++ {
			if len(result) == 10 {
				break
			}

			item, ok := items[i].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]",
					requestUrl.String(),
					i,
				)
			}

			id, ok := item["id"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['id']",
					requestUrl.String(),
					i,
				)
			}

			video, ok := id["videoId"].(string)
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['videoId']",
					requestUrl.String(),
					i,
				)
			}

			videoMetadata, err := youtube.GetVideoMetadata(video)
			if err != nil {
				return nil, fmt.Errorf(
					"error in finding metadata for video %s",
					video,
				)
			}
			result = append(result, videoMetadata)
		}
		return &VideoList{
			Count:  len(result),
			Videos: result,
		}, nil
	} else {
		items, ok := body["items"].([]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"error in YouTube Data API response from %s. Key 'items' not found",
				requestUrl.String(),
			)
		}
		if len(items) == 0 {
			return nil, fmt.Errorf(
				"no videos found for channel with channel id %s",
				channelId,
			)
		}

		for i := 0; i < len(items); i++ {
			item, ok := items[i].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]",
					requestUrl.String(),
					i,
				)
			}

			id, ok := item["id"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['id']",
					requestUrl.String(),
					i,
				)
			}

			video, ok := id["videoId"].(string)
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['videoId']",
					requestUrl.String(),
					i,
				)
			}

			if video == videoId {
				break
			}

			videoMetadata, err := youtube.GetVideoMetadata(video)
			if err != nil {
				return nil, fmt.Errorf(
					"error in finding metadata for video %s",
					video,
				)
			}
			result = append(result, videoMetadata)
		}

		// fetching older videos
		previousPageToken, previousPageExists := body["prevPageToken"].(string)
		if !previousPageExists {
			return &VideoList{
				Count:  len(result),
				Videos: result,
			}, nil
		}

		q := requestUrl.Query()
		q.Set("key", youtube.apiKey)
		q.Set("channelId", channelId)
		q.Set("maxResults", "10")
		q.Set("type", "video")
		q.Set("order", "date")
		q.Set("pageToken", previousPageToken)

		requestUrl.RawQuery = q.Encode()

		res, err = http.Get(requestUrl.String())
		if err != nil {
			return nil, err
		}

		if res.StatusCode != 200 {
			return nil, fmt.Errorf(
				"call to YouTube API endpoint %s failed with message %s",
				endpoint,
				res.Status,
			)
		}

		body = make(map[string]interface{})

		json.NewDecoder(res.Body).Decode(&body)

		res.Body.Close()

		items, ok = body["items"].([]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"error in YouTube Data API response from %s. Key 'items' not found",
				requestUrl.String(),
			)
		}
		if len(items) == 0 {
			return nil, fmt.Errorf(
				"no videos found for channel with channel id %s",
				channelId,
			)
		}

		for i := len(items) - 1; i >= 0; i-- {
			if len(result) == 10 {
				break
			}

			item, ok := items[i].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]",
					requestUrl.String(),
					i,
				)
			}

			id, ok := item["id"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['id']",
					requestUrl.String(),
					i,
				)
			}

			video, ok := id["videoId"].(string)
			if !ok {
				return nil, fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['videoId']",
					requestUrl.String(),
					i,
				)
			}

			videoMetadata, err := youtube.GetVideoMetadata(video)
			if err != nil {
				return nil, fmt.Errorf(
					"error in finding metadata for video %s",
					video,
				)
			}
			result = append(result, videoMetadata)
		}

		return &VideoList{
			Count:  len(result),
			Videos: result,
		}, nil
	}
}
