# GitHub Actions 权限配置指南

## 问题描述

GitHub Actions bot 没有权限推送到仓库，出现 403 错误。

## 解决方案

### 方案一：配置仓库权限设置（推荐）

1. 进入你的 GitHub 仓库
2. 点击 **Settings** → **Actions** → **General**
3. 向下滚动到 **Workflow permissions** 部分
4. 选择 **Read and write permissions**
5. 勾选 **Allow GitHub Actions to create and approve pull requests**（可选）
6. 点击 **Save**

### 方案二：使用 Personal Access Token (PAT)

如果方案一不起作用，可以使用个人访问令牌：

#### 步骤 1：创建 Personal Access Token

1. 进入 GitHub 账号设置：https://github.com/settings/profile
2. 点击左侧菜单的 **Developer settings**
3. 点击 **Personal access tokens** → **Tokens (classic)**
4. 点击 **Generate new token** → **Generate new token (classic)**
5. 设置以下选项：
   - **Note**: `GitHub Actions Node Checker`（或其他描述）
   - **Expiration**: 选择合适的过期时间（建议 90 天或无过期）
   - **Scopes**: 勾选 `repo`（完整仓库权限）
6. 点击 **Generate token**
7. **重要**：复制生成的 token（只显示一次！）

#### 步骤 2：添加到仓库 Secrets

1. 进入你的仓库
2. 点击 **Settings** → **Secrets and variables** → **Actions**
3. 点击 **New repository secret**
4. **Name**: `PERSONAL_ACCESS_TOKEN`
5. **Value**: 粘贴刚才复制的 token
6. 点击 **Add secret**

#### 步骤 3：更新 workflow 文件

修改 `.github/workflows/node-check.yml` 中的检出代码部分：

```yaml
- name: 检出代码
  uses: actions/checkout@v4
  with:
    fetch-depth: 0
    token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}  # 使用 PAT 而不是 GITHUB_TOKEN
```

## 已更新的 workflow 配置

我已经在 workflow 文件中添加了 `permissions: contents: write`，这应该能解决大部分权限问题。

## 验证配置

配置完成后，可以手动触发 workflow 测试：

1. 进入仓库的 **Actions** 标签页
2. 选择 "每日节点验证" 工作流
3. 点击 **Run workflow**
4. 选择分支，点击 **Run workflow**

查看执行日志确认是否成功！

## 常见问题

### Q: 为什么需要 write 权限？
A: GitHub Actions 默认只有读取权限，需要显式授予写入权限才能提交代码更改。

### Q: 使用 PAT 安全吗？
A: 是的，只要 token 只用于这个仓库且权限范围合理。建议定期轮换 token。

### Q: 还有其他方法吗？
A: 可以使用部署密钥（Deploy Key），但配置更复杂。PAT 是最简单的方案。
