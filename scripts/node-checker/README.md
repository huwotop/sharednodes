# 每日节点自动验证工具

这是一个基于 GitHub Actions 的自动化工具，用于每天自动验证订阅链接中的节点，并将有效节点保存到项目的 `jiedian` 文件夹中。

## 📁 文件结构

```
新项目/
├── .github/
│   └── workflows/
│       └── node-check.yml          # GitHub Actions 工作流
├── scripts/
│   └── node-checker/
│       ├── main.go                  # 节点验证主程序
│       ├── go.mod                   # Go 模块文件
│       └── README.md                # 本文档
└── jiedian/                         # 保存有效节点的目录
    ├── nodes-latest.yaml            # 最新有效节点
    └── nodes-2026-03-09.yaml      # 按日期保存的节点
```

## 🚀 使用步骤

### 1. 创建新仓库

将这个"新项目"文件夹的内容上传到你自己的 GitHub 仓库。

### 2. 配置 Secrets

在你的 GitHub 仓库中设置以下 Secrets：

1. 进入仓库的 **Settings** → **Secrets and variables** → **Actions**
2. 点击 **New repository secret**
3. 添加以下 Secret：

| 名称 | 说明 | 示例 |
|------|------|------|
| `SUBSCRIPTION_URL` | 你的订阅链接 | `https://example.com/subscription` |

### 3. 启用 Workflow

1. 进入仓库的 **Actions** 标签页
2. 找到 "每日节点验证" 工作流
3. 点击 **Enable workflow** 启用它

### 4. 手动测试（可选）

你可以手动触发工作流来测试：

1. 进入 **Actions** → "每日节点验证"
2. 点击 **Run workflow**
3. 选择分支，点击 **Run workflow**

## ⚙️ 配置选项

你可以通过修改环境变量来调整工具行为：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `SUBSCRIPTION_URL` | 订阅地址（必需） | - |
| `TEST_URL` | 节点测试 URL | `https://www.gstatic.com/generate_204` |
| `TIMEOUT` | 单个节点超时时间（秒） | `10` |
| `THREAD` | 并发检测数量 | `50` |
| `OUTPUT_DIR` | 输出目录 | `../../jiedian` |

要修改这些配置，编辑 `.github/workflows/node-check.yml` 文件：

```yaml
- name: 运行节点验证
  run: |
    cd scripts/node-checker
    go run main.go
  env:
    SUBSCRIPTION_URL: ${{ secrets.SUBSCRIPTION_URL }}
    TEST_URL: "https://your-test-url.com"  # 添加自定义测试 URL
    TIMEOUT: "15"                            # 修改超时时间
    THREAD: "100"                            # 修改并发数
```

## 📊 输出文件

工具会在 `jiedian` 目录下生成两个文件：

1. **`nodes-latest.yaml`** - 始终指向最新的有效节点
2. **`nodes-YYYY-MM-DD.yaml`** - 按日期保存的历史节点记录

## 🔧 修改执行时间

默认情况下，工具会在 **每天 UTC 0:00**（北京时间 8:00）执行。

要修改执行时间，编辑 `.github/workflows/node-check.yml` 中的 `cron` 表达式：

```yaml
on:
  schedule:
    - cron: '0 0 * * *'  # 修改这里
```

Cron 表达式格式：`分 时 日 月 周`

**示例**：
- 每天 UTC 2:00（北京时间 10:00）：`'0 2 * * *'`
- 每周一 UTC 0:00：`'0 0 * * 1'`
- 每小时执行一次：`'0 * * * *'`

> ⚠️ 注意：GitHub Actions 使用 UTC 时间，北京时间 = UTC + 8 小时

## 🛠️ 本地测试

你也可以在本地测试这个工具：

```bash
# 进入目录
cd scripts/node-checker

# 设置环境变量
export SUBSCRIPTION_URL="https://your-subscription-url.com"

# 运行
go run main.go
```

## 📝 注意事项

1. **GitHub Actions 配额**：免费账户每月有 2000 分钟的执行时间
2. **提交权限**：确保你的 GitHub Token 有提交权限
3. **订阅格式**：工具支持 Clash/Mihomo 格式的订阅
4. **网络限制**：GitHub Actions 环境可能无法访问某些节点

## 🐛 故障排除

### 工作流没有自动触发

- 检查 cron 表达式是否正确
- 确保工作流已启用
- 查看 Actions 日志了解详细信息

### 节点检测全部失败

- 检查订阅链接是否有效
- 尝试修改 `TEST_URL`
- 检查网络连接

### 没有提交更改

- 确保 `jiedian` 目录存在
- 检查 GitHub Token 权限
- 查看 Actions 日志中的错误信息

## 📄 许可证

与主项目保持一致。
