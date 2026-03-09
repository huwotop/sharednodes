package fetcher

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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
	var result [][]byte

	var clashConfig struct {
		Proxies []map[string]any `yaml:"proxies"`
	}

	if err := yaml.Unmarshal(content, &clashConfig); err == nil && len(clashConfig.Proxies) > 0 {
		for _, proxy := range clashConfig.Proxies {
			nodeYaml, err := yaml.Marshal(proxy)
			if err == nil {
				result = append(result, nodeYaml)
			}
		}
		return result
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) > 1 {
		lines = lines[1:]
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if len(line) > 4 {
				line = line[4:]
			}
			result = append(result, line)
		}
		return result
	}

	contentStr := string(content)
	linesStr := strings.Split(contentStr, "\n")
	for _, line := range linesStr {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, []byte(line))
	}

	return result
}
