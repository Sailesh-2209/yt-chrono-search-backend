package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"
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
	PublishedAt      string
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

type PlaylistVideo struct {
	VideoId     string
	PublishedAt string
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

	publishedAt, ok := snippet["publishedAt"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'publishedAt' not found in items[0]['snippet']",
			url.String(),
		)
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
		PublishedAt:      publishedAt,
		ChannelId:        channelId,
		ChannelTitle:     channelTitle,
		ChannelThumbnail: channelThumbnailUrl,
		ChannelCustomUrl: customUrl,
		SubscriberCount:  subscriberCount,
		VideoCount:       videoCount,
	}, nil
}

func (youtube *YouTube) GetChannelVideos(
	channelId string, videoId string,
) (*VideoList, error) {
	playlistId, err := youtube.GetUploadsPlaylist(channelId)
	if err != nil {
		return nil, err
	}

	totalResults, err := youtube.GetPlaylistVideoCount(playlistId)
	if err != nil {
		return nil, err
	}

	videos := make([]PlaylistVideo, totalResults)

	chPageTokens := make(chan string, 20)
	chPageTokensErrors := make(chan error, 10)
	chPageTokensDone := make(chan bool)
	chVideos := make(chan PlaylistVideo, 100)

	chPageTokens <- ""

	stop := false

	go func() {
		for !stop {
			select {
			case pageToken := <-chPageTokens:
				youtube.GetPageVideos(
					playlistId,
					pageToken,
					chPageTokens,
					chPageTokensErrors,
					chPageTokensDone,
					chVideos,
				)
			case <-chPageTokensDone:
				select {
				case pageToken := <-chPageTokens:
					youtube.GetPageVideos(
						playlistId,
						pageToken,
						chPageTokens,
						chPageTokensErrors,
						chPageTokensDone,
						chVideos,
					)
				default:
					stop = true
				}
			default:
				stop = true
			}
		}
	}()

	for i := 0; i < totalResults; i++ {
		videos[i] = <-chVideos
	}

	slices.SortFunc(videos,
		func(a, b PlaylistVideo) int {
			timeFormat := "2006-01-02T15:04:05Z"
			timeA, _ := time.Parse(timeFormat, a.PublishedAt)
			timeB, _ := time.Parse(timeFormat, b.PublishedAt)
			if timeA.After(timeB) {
				return 1
			} else if timeA.Before(timeB) {
				return -1
			} else {
				return 0
			}
		},
	)

	ind := -1

	for i, vid := range videos {
		if vid.VideoId == videoId {
			ind = i
			break
		}
	}

	if ind == -1 {
		return nil, fmt.Errorf("video not found in uploads playlist")
	}

	chRequiredVideos := make(chan *VideoMetadata, 21)

	for j := ind - 10; j <= ind+10; j++ {
		go func(j int) {
			if j < len(videos) && j >= 0 {
				data, err := youtube.GetVideoMetadata(videos[j].VideoId)
				if err != nil {
					fmt.Fprintf(
						os.Stderr,
						"Error finding metadata for %s: %s",
						videos[j].VideoId,
						err,
					)
					chRequiredVideos <- nil
				} else {
					chRequiredVideos <- data
				}
			} else {
				chRequiredVideos <- nil
			}
		}(j)
	}

	requiredVideos := []*VideoMetadata{}

	for i := 0; i < 21; i++ {
		video := <-chRequiredVideos
		if video != nil {
			requiredVideos = append(requiredVideos, video)
		}
	}

	slices.SortFunc(requiredVideos,
		func(a, b *VideoMetadata) int {
			timeFormat := "2006-01-02T15:04:05Z"
			timeA, _ := time.Parse(timeFormat, a.PublishedAt)
			timeB, _ := time.Parse(timeFormat, b.PublishedAt)
			if timeA.After(timeB) {
				return 1
			} else if timeA.Before(timeB) {
				return -1
			} else {
				return 0
			}
		},
	)

	return &VideoList{
		Count:  len(requiredVideos),
		Videos: requiredVideos,
	}, nil
}
