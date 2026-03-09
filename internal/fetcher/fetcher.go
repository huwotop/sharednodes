package fetcher

import (
	"io"
	"net/http"
	"strings"
	"time"
)

func FetchSubscription(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func ParseNodes(content []byte) [][]byte {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var nodes [][]byte

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		nodes = append(nodes, []byte(line))
	}

	return nodes
}
