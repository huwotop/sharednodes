package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/bestruirui/bestsub-action/internal/checker"
	"github.com/bestruirui/bestsub-action/internal/config"
	"github.com/bestruirui/bestsub-action/internal/converter"
	"github.com/bestruirui/bestsub-action/internal/fetcher"
	"github.com/bestruirui/bestsub-action/internal/model"
)

func main() {
	fmt.Println("=== BestSub 自动节点筛选工具")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	nodePool := model.NewNodePool(cfg.Check.NodePoolSize)

	for _, url := range cfg.Subscription.URLs {
		fmt.Printf("正在获取订阅: %s\n", url)
		content, err := fetcher.FetchSubscription(url)
		if err != nil {
			fmt.Printf("获取订阅失败: %v\n", err)
			continue
		}

		nodesRaw := fetcher.ParseNodes(content)
		fmt.Printf("解析到 %d 个节点\n", len(nodesRaw))

		for _, raw := range nodesRaw {
			var uk model.UniqueKey
			if err := yaml.Unmarshal(raw, &uk); err != nil {
				continue
			}
			node := model.Node{
				Raw:       raw,
				UniqueKey: uk.Gen(),
				Info: &model.NodeInfo{
					AliveStatus: 0,
				},
			}
			nodePool.Add(node)
		}
	}

	fmt.Printf("节点池总共有 %d 个节点\n", nodePool.Size())

	fmt.Println("开始检测节点存活状态...")
	aliveChecker := checker.NewAliveChecker(
		cfg.Check.AliveTestURL,
		cfg.Check.AliveTestStatusCode,
		cfg.Check.AliveTimeout,
		cfg.Check.MaxThreads,
	)

	nodes := nodePool.GetAll()
	aliveCount, deadCount, avgDelay := aliveChecker.Check(nodes)
	fmt.Printf("检测完成: 存活 %d, 死亡 %d, 平均延迟 %dms\n", aliveCount, deadCount, avgDelay)

	filteredNodes := checker.FilterAlive(nodes, cfg.Filter.MaxDelay)
	fmt.Printf("筛选后剩余 %d 个节点\n", len(filteredNodes))

	var nodeRaws [][]byte
	for _, node := range filteredNodes {
		nodeRaws = append(nodeRaws, node.Raw)
	}

	links, err := converter.NodesToShareLinks(nodeRaws)
	if err != nil {
		fmt.Printf("转换节点链接失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("成功转换 %d 个分享链接\n", len(links))

	output := converter.GenerateShareLinksFile(links)
	err = os.WriteFile(cfg.Output.FileName, output, 0644)
	if err != nil {
		fmt.Printf("写入输出文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("节点已保存到 %s\n", cfg.Output.FileName)
}
