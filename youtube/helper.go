package youtube

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func parseVideoId(idOrUrl string) (string, error) {
	oIdOrUrl := idOrUrl

	if !strings.HasPrefix(idOrUrl, "http://") &&
		!strings.HasPrefix(idOrUrl, "https://") {
		return idOrUrl, nil
	}

	idOrUrl = strings.TrimPrefix(idOrUrl, "http://")
	idOrUrl = strings.TrimPrefix(idOrUrl, "https://")

	if strings.HasPrefix(idOrUrl, "www.youtube.com") ||
		strings.HasPrefix(idOrUrl, "youtube.com") {
		pattern := `^(?:www\.)?youtube\.com/watch/?\?(?:[^&]*&)*v=([a-zA-Z0-9_!-]+)(?:&[^&]*)*$`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(idOrUrl)

		if len(match) > 1 {
			return match[1], nil
		} else {
			return "", fmt.Errorf("video ID not found in URL: %s", oIdOrUrl)
		}
	} else if strings.HasPrefix(idOrUrl, "youtu.be") {
		pattern := `^youtu.be/([^?]*)\??.*$`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(idOrUrl)

		if len(match) > 1 {
			return match[1], nil
		} else {
			return "", fmt.Errorf("video ID not found in URL: %s", oIdOrUrl)
		}
	} else {
		return "", fmt.Errorf("unrecognized youtube URL format: %s", oIdOrUrl)
	}
}

func (youtube *YouTube) GetUploadsPlaylist(channelId string) (string, error) {
	const endpoint = "channels/"

	requestUrl, err := url.Parse(BaseUrl + endpoint)
	if err != nil {
		return "", err
	}

	q := requestUrl.Query()
	q.Set("key", youtube.apiKey)
	q.Set("id", channelId)
	q.Set("part", "contentDetails")
	requestUrl.RawQuery = q.Encode()

	res, err := http.Get(requestUrl.String())
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf(
			"call to YouTube API endpoint %s failed with message %s",
			endpoint,
			res.Status,
		)
	}

	body := make(map[string]interface{})

	json.NewDecoder(res.Body).Decode(&body)

	res.Body.Close()

	items, ok := body["items"].([]interface{})
	if !ok {
		return "", fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'items' not found",
			requestUrl.String(),
		)
	}
	if len(items) == 0 {
		return "", fmt.Errorf(
			"no items found in YouTube Data API response from %s",
			requestUrl.String(),
		)
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(
			"error in YouTube Data API response from %s. Cannot access items[0]['contentDetails']",
			requestUrl.String(),
		)
	}

	contentDetails, ok := item["contentDetails"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(
			"error in YouTube Data API response from %s. Cannot access items[0]['contentDetails']",
			requestUrl.String(),
		)
	}

	relatedPlaylists, ok := contentDetails["relatedPlaylists"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf(
			"error in YouTube Data API response from %s. Cannot access items[0]['contentDetails']['relatedPlaylists']",
			requestUrl.String(),
		)
	}

	uploadsPlaylist, ok := relatedPlaylists["uploads"].(string)
	if !ok {
		return "", fmt.Errorf(
			"error in YouTube Data API response from %s. Cannot access items[0]['contentDetails']['relatedPlaylists']['uploads']",
			requestUrl.String(),
		)
	}

	return uploadsPlaylist, nil
}

func (youtube *YouTube) GetPlaylistVideoCount(playlistId string) (int, error) {
	const endpoint = "playlistItems/"

	requestUrl, err := url.Parse(BaseUrl + endpoint)
	if err != nil {
		return 0, err
	}

	q := requestUrl.Query()
	q.Set("key", youtube.apiKey)
	q.Set("maxResults", "1")
	q.Set("part", "contentDetails")
	q.Set("playlistId", playlistId)
	requestUrl.RawQuery = q.Encode()

	res, err := http.Get(requestUrl.String())
	if err != nil {
		return 0, err
	}

	if res.StatusCode != 200 {
		return 0, fmt.Errorf(
			"call to YouTube API endpoint %s failed with message %s. Request URL: %s",
			endpoint,
			res.Status,
			requestUrl.String(),
		)
	}

	body := make(map[string]interface{})

	json.NewDecoder(res.Body).Decode(&body)

	res.Body.Close()

	pageInfo, ok := body["pageInfo"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf(
			"error in YouTube Data API response from %s. Key 'pageInfo' not found",
			requestUrl.String(),
		)
	}

	totalResults, ok := pageInfo["totalResults"].(float64)
	if !ok {
		return 0, fmt.Errorf(
			"error in YouTube Data API response from %s. Cannot access pageInfo['totalResults']",
			requestUrl.String(),
		)
	}

	return int(math.Round(totalResults)), nil
}

func (youtube *YouTube) GetPageVideos(
	playlistId string,
	pageToken string,
	chPageTokens chan<- string,
	chPageTokensErrors chan<- error,
	chPageTokensDone chan<- bool,
	chVideos chan<- PlaylistVideo,
) {
	const endpoint = "playlistItems/"

	requestUrl, err := url.Parse(BaseUrl + endpoint)
	if err != nil {
		select {
		case chPageTokensErrors <- err:
		default:
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
	}

	q := requestUrl.Query()
	q.Set("key", youtube.apiKey)
	q.Set("maxResults", "50")
	q.Set("part", "contentDetails")
	q.Set("playlistId", playlistId)
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}
	requestUrl.RawQuery = q.Encode()

	res, err := http.Get(requestUrl.String())
	if err != nil {
		select {
		case chPageTokensErrors <- err:
		default:
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
	}

	if res.StatusCode != 200 {
		select {
		case chPageTokensErrors <- err:
		default:
			fmt.Fprintf(
				os.Stderr,
				"call to YouTube API endpoint %s failed with message %s",
				endpoint,
				res.Status,
			)
			return
		}
	}

	body := make(map[string]interface{})

	json.NewDecoder(res.Body).Decode(&body)

	res.Body.Close()

	nextPageToken, nextPageExists := body["nextPageToken"].(string)
	if !nextPageExists {
		select {
		case chPageTokensDone <- true:
		default:
			fmt.Fprintf(os.Stderr, "Attempting to write to unbuffered channel twice.")
		}
	} else {
		chPageTokens <- nextPageToken
	}

	items, ok := body["items"].([]interface{})
	if !ok {
		select {
		case chPageTokensErrors <- err:
		default:
			fmt.Fprintf(
				os.Stderr,
				"error in YouTube Data API response from %s. Key 'items' not found",
				requestUrl.String(),
			)
			return
		}
	}

	for i := 0; i < len(items); i++ {
		item, ok := items[i].(map[string]interface{})
		if !ok {
			select {
			case chPageTokensErrors <- err:
			default:
				fmt.Fprintf(
					os.Stderr,
					"error in YouTube Data API response from %s. Key items[%d] not found",
					requestUrl.String(),
					i,
				)
				return
			}
		}

		contentDetails, ok := item["contentDetails"].(map[string]interface{})
		if !ok {
			select {
			case chPageTokensErrors <- err:
			default:
				fmt.Fprintf(
					os.Stderr,
					"error in YouTube Data API response from %s. Key items[%d]['contentDetails'] not found",
					requestUrl.String(),
					i,
				)
				return
			}
		}

		videoId, ok := contentDetails["videoId"].(string)
		if !ok {
			select {
			case chPageTokensErrors <- err:
			default:
				fmt.Fprintf(
					os.Stderr,
					"error in YouTube Data API response from %s. Key items[%d]['contentDetails']['videoId'] not found",
					requestUrl.String(),
					i,
				)
				return
			}
		}

		publishedAt, ok := contentDetails["videoPublishedAt"].(string)
		if !ok {
			select {
			case chPageTokensErrors <- err:
			default:
				fmt.Fprintf(
					os.Stderr,
					"error in YouTube Data API response from %s. Key items[%d]['contentDetails']['videoPublishedAt'] not found",
					requestUrl.String(),
					i,
				)
				return
			}
		}

		chVideos <- PlaylistVideo{
			VideoId:     videoId,
			PublishedAt: publishedAt,
		}
	}
}
