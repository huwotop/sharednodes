package fetcher

import (
	"bytes"
	"fmt"
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
		fmt.Printf("检测到 Clash 配置文件，包含 %d 个节点\n", len(clashConfig.Proxies))
		for i, proxy := range clashConfig.Proxies {
			nodeYaml, err := yaml.Marshal(proxy)
			if err == nil {
				result = append(result, nodeYaml)
				if i < 3 {
					fmt.Printf("  节点 %d: %v\n", i+1, string(nodeYaml))
				}
			}
		}
		return result
	}

	fmt.Println("尝试按行解析...")
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
