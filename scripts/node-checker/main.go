package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

type Node struct {
	Name string                 `yaml:"name"`
	Raw  map[string]interface{} `yaml:"-"`
}

type Config struct {
	SubscriptionURL string
	TestURL         string
	Timeout         int
	Thread          int
	OutputDir       string
}

func loadConfig() Config {
	return Config{
		SubscriptionURL: getEnv("SUBSCRIPTION_URL", ""),
		TestURL:         getEnv("TEST_URL", "https://www.gstatic.com/generate_204"),
		Timeout:         getEnvInt("TIMEOUT", 10),
		Thread:          getEnvInt("THREAD", 50),
		OutputDir:       getEnv("OUTPUT_DIR", "../../jiedian"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		_, err := fmt.Sscanf(value, "%d", &result)
		if err == nil {
			return result
		}
	}
	return defaultValue
}

func fetchSubscription(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "BestSub-Node-Checker/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func parseNodes(content []byte) ([]Node, error) {
	var nodes []Node
	lines := bytes.Split(content, []byte("\n"))

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, []byte("proxies:")) {
			continue
		}

		if bytes.HasPrefix(line, []byte("- ")) {
			line = line[2:]
		}

		var node Node
		var raw map[string]interface{}
		if err := yaml.Unmarshal(line, &raw); err != nil {
			continue
		}

		if name, ok := raw["name"].(string); ok {
			node.Name = name
			node.Raw = raw
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func detectNode(ctx context.Context, node Node, testURL string, timeout int) bool {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 204 || resp.StatusCode == 200
}

func checkNodes(nodes []Node, config Config) ([]Node, error) {
	var aliveNodes []Node
	var aliveCount, deadCount int64

	threads := config.Thread
	if threads <= 0 || threads > len(nodes) {
		threads = len(nodes)
	}

	sem := make(chan struct{}, threads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	ctx := context.Background()

	for _, node := range nodes {
		sem <- struct{}{}
		wg.Add(1)

		go func(n Node) {
			defer func() {
				<-sem
				wg.Done()
			}()

			if detectNode(ctx, n, config.TestURL, config.Timeout) {
				atomic.AddInt64(&aliveCount, 1)
				mu.Lock()
				aliveNodes = append(aliveNodes, n)
				mu.Unlock()
				fmt.Printf("✓ %s 存活\n", n.Name)
			} else {
				atomic.AddInt64(&deadCount, 1)
				fmt.Printf("✗ %s 失效\n", n.Name)
			}
		}(node)
	}

	wg.Wait()

	fmt.Printf("\n检测完成: 存活 %d, 失效 %d\n", aliveCount, deadCount)
	return aliveNodes, nil
}

func saveNodes(nodes []Node, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString("proxies:\n")

	for _, node := range nodes {
		data, err := yaml.Marshal(node.Raw)
		if err != nil {
			continue
		}
		lines := bytes.Split(data, []byte("\n"))
		for _, line := range lines {
			if len(line) > 0 {
				buf.WriteString("  - ")
				buf.Write(line)
				buf.WriteString("\n")
			}
		}
	}

	timestamp := time.Now().Format("2006-01-02")
	filename := filepath.Join(outputDir, fmt.Sprintf("nodes-%s.yaml", timestamp))

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}

	latestFilename := filepath.Join(outputDir, "nodes-latest.yaml")
	if err := os.WriteFile(latestFilename, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("已保存 %d 个有效节点到 %s\n", len(nodes), filename)
	return nil
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  BestSub 节点验证工具")
	fmt.Println("========================================")

	config := loadConfig()

	if config.SubscriptionURL == "" {
		fmt.Println("错误: 请设置 SUBSCRIPTION_URL 环境变量")
		os.Exit(1)
	}

	fmt.Printf("订阅地址: %s\n", config.SubscriptionURL)
	fmt.Printf("测试地址: %s\n", config.TestURL)
	fmt.Printf("超时时间: %d秒\n", config.Timeout)
	fmt.Printf("并发数: %d\n", config.Thread)
	fmt.Printf("输出目录: %s\n", config.OutputDir)
	fmt.Println()

	fmt.Println("正在获取订阅...")
	content, err := fetchSubscription(config.SubscriptionURL)
	if err != nil {
		fmt.Printf("获取订阅失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("获取成功，数据大小: %d bytes\n", len(content))

	fmt.Println("\n正在解析节点...")
	nodes, err := parseNodes(content)
	if err != nil {
		fmt.Printf("解析节点失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("解析到 %d 个节点\n", len(nodes))

	if len(nodes) == 0 {
		fmt.Println("没有找到节点")
		os.Exit(0)
	}

	fmt.Println("\n正在检测节点...")
	aliveNodes, err := checkNodes(nodes, config)
	if err != nil {
		fmt.Printf("检测节点失败: %v\n", err)
		os.Exit(1)
	}

	if len(aliveNodes) == 0 {
		fmt.Println("没有存活的节点")
		os.Exit(0)
	}

	fmt.Println("\n正在保存节点...")
	if err := saveNodes(aliveNodes, config.OutputDir); err != nil {
		fmt.Printf("保存节点失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n完成!")
}
