package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func findPageToken(
	videoId string, channelId string, apiKey string,
) (string, error) {
	const endpoint = "search/"

	requestUrl, err := url.Parse(BaseUrl + endpoint)
	if err != nil {
		return "", fmt.Errorf("error in forming URL for endpoint %s", endpoint)
	}

	nextPageToken, nextPageExists := "", false

	returnPrevPageToken := false

	for {
		q := requestUrl.Query()
		q.Set("key", apiKey)
		q.Set("channelId", channelId)
		q.Set("maxResults", "10")
		q.Set("type", "video")
		q.Set("order", "date")
		if nextPageExists {
			q.Set("pageToken", nextPageToken)
		}
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

		if returnPrevPageToken {
			token, ok := body["prevPageToken"].(string)
			if !ok {
				return "", fmt.Errorf("unable to find page token")
			}
			return token, nil
		}

		items, ok := body["items"].([]interface{})
		if !ok {
			return "", fmt.Errorf(
				"error in YouTube Data API response from %s. Key 'items' not found",
				requestUrl.String(),
			)
		}
		if len(items) == 0 {
			return "", fmt.Errorf(
				"no videos found for channel with channel id %s",
				channelId,
			)
		}

		for i := 0; i < len(items); i++ {
			item, ok := items[i].(map[string]interface{})
			if !ok {
				return "", fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]",
					requestUrl.String(),
					i,
				)
			}

			id, ok := item["id"].(map[string]interface{})
			if !ok {
				return "", fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['id']",
					requestUrl.String(),
					i,
				)
			}

			video, ok := id["videoId"].(string)
			if !ok {
				return "", fmt.Errorf(
					"error in YouTube Data API response from %s. Cannot access items[%d]['videoId']",
					requestUrl.String(),
					i,
				)
			}

			if video == videoId {
				if nextPageToken == "" {
					returnPrevPageToken = true
				} else {
					return nextPageToken, nil
				}
			}
		}

		nextPageToken, nextPageExists = body["nextPageToken"].(string)
		if !nextPageExists {
			break
		}
	}

	return "", fmt.Errorf(
		"video %s not found in channel %s",
		videoId,
		channelId,
	)
}
