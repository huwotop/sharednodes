# BestSub GitHub Action 自动节点筛选

这是一个简化版的 BestSub，专门用于 GitHub Action 自动筛选有效节点。

## 功能特性

- 🚀 自动从订阅链接获取节点
- ✅ 检测节点存活状态
- ⚡ 并发检测，高性能
- 📊 按延迟筛选节点
- 🔄 每天自动运行（可配置）

## 使用方法

### 1. 配置订阅链接

编辑 `config.yaml` 文件，添加你的订阅链接：

```yaml
subscription:
  urls:
    - "https://your-subscribe-url-1"
    - "https://your-subscribe-url-2"
```

### 2. 调整筛选参数

在 `config.yaml` 中可以调整：

- `alive_timeout`: 单个节点超时时间（秒）
- `max_threads`: 并发检测线程数
- `max_delay`: 最大允许延迟（毫秒）

### 3. 本地运行

```bash
# 构建项目
go build -o bestsub-action .

# 运行
./bestsub-action
```

## GitHub Action 配置

项目已配置好 GitHub Action，会自动：

- **每天 00:00 UTC（北京时间 08:00）** 自动运行
- 筛选节点后自动提交到仓库
- 同时上传 Artifact 供下载

### 手动触发

你也可以在 GitHub 仓库的 Actions 页面手动触发工作流。

## 输出

筛选后的节点会保存为 `filtered_nodes.yaml`，格式为 Clash 兼容的 YAML 格式。

## 注意事项

⚠️ 本项目仅供学习和研究使用。使用本软件时，请遵守当地法律法规。
