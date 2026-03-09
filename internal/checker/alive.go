package checker

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bestruirui/bestsub-action/internal/mihomo"
	"github.com/bestruirui/bestsub-action/internal/model"
	"github.com/panjf2000/ants/v2"
)

type AliveChecker struct {
	TestURL      string
	ExpectedCode int
	Timeout      int
	MaxThreads   int
}

func NewAliveChecker(testURL string, expectedCode int, timeout int, maxThreads int) *AliveChecker {
	return &AliveChecker{
		TestURL:      testURL,
		ExpectedCode: expectedCode,
		Timeout:      timeout,
		MaxThreads:   maxThreads,
	}
}

func (c *AliveChecker) Check(nodes []model.Node) (aliveCount, deadCount int64, avgDelay int64) {
	var wg sync.WaitGroup
	var totalDelay int64

	pool, _ := ants.NewPool(c.MaxThreads)
	defer pool.Release()

	for i := range nodes {
		wg.Add(1)
		node := &nodes[i]
		pool.Submit(func() {
			defer wg.Done()
			delay, alive := c.checkNode(node)
			if alive {
				atomic.AddInt64(&aliveCount, 1)
				atomic.AddInt64(&totalDelay, int64(delay))
				node.Info.Delay = delay
				node.Info.SetAliveStatus(model.Alive, true)
			} else {
				atomic.AddInt64(&deadCount, 1)
				node.Info.SetAliveStatus(model.Alive, false)
			}
		})
	}

	wg.Wait()

	if aliveCount > 0 {
		avgDelay = totalDelay / aliveCount
	}

	return
}

func (c *AliveChecker) checkNode(node *model.Node) (uint16, bool) {
	var raw map[string]any
	if err := yaml.Unmarshal(node.Raw, &raw); err != nil {
		return 0, false
	}

	client := mihomo.Proxy(raw)
	if client == nil {
		return 0, false
	}
	defer client.Release()
	client.Timeout = time.Duration(c.Timeout) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.TestURL, nil)
	if err != nil {
		return 0, false
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()

	delay := uint16(time.Since(start).Milliseconds())
	return delay, resp.StatusCode == c.ExpectedCode
}

func FilterAlive(nodes []model.Node, maxDelay uint16) []model.Node {
	var result []model.Node
	for _, node := range nodes {
		if node.Info.AliveStatus&model.Alive != 0 {
			if maxDelay == 0 || node.Info.Delay <= maxDelay {
				result = append(result, node)
			}
		}
	}
	return result
}
