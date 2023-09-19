package youtube

import (
	"fmt"
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
