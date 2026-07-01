# Git 工作流与里程碑管理

> 本文档说明本项目的 Git 使用约定：什么该推、什么不推、如何用 **tag** 标记每个学习阶段，以及如何查看 / 回退历史。

远程仓库：[agent_backend](https://github.com/T-DWAG/agent_backend)

---

## 一、项目里的两类内容

| 类型 | 路径示例 | 是否推送 | 说明 |
|------|----------|----------|------|
| 教程原文 | `01-开发环境准备/`、`02-项目搭建/` | ❌ | 老师发的资料，仅本地阅读 |
| 学习笔记 | `md/ch2-项目搭建.md` | ✅ | 自己整理的章节总结 |
| 项目代码 | `app/`、`common/`、`core/`、`model/` | ✅ | 实际开发代码 |
| 中间件配置 | `docker/docker-compose.yml` 等 | ✅ | 不含运行时数据 |
| 敏感 / 运行时 | `docker/.env`、`docker/data/` | ❌ | 密码和数据库文件 |

`.gitignore` 关键规则：

```gitignore
/[0-9]*/          # 仅忽略根目录数字开头的教程文件夹
docker/.env
docker/data/
```

> 笔记放在 `md/` 下，命名用 `ch1-`、`ch2-` 前缀，**不要**用 `02-` 开头（避免和 ignore 规则混淆）。

---

## 二、时间步 = commit + tag

每完成一个学习阶段：

1. 提交代码和笔记 → 产生一个 **commit**
2. 打 **tag** 标记里程碑 → 方便以后查找，不用记 hash

### 当前里程碑

| Tag | Commit | 说明 |
|-----|--------|------|
| `ch1-env` | `d3ba30f` | init: 开发环境准备 |
| `ch2-scaffold` | `4624fc7` | feat: 项目搭建 |

查看历史：

```powershell
cd E:\data\code\eino\project\backend
git log --oneline --decorate
git tag -l
```

---

## 三、每章完成后的推送流程

```powershell
cd E:\data\code\eino\project\backend

# 1. 只 add 该推的内容（不要用 git add . 误加 docker/data）
git add .gitignore go.work app/ common/ core/ model/ md/
git status

# 2. 确认 status 里没有：
#    - 01-xxx / 02-xxx 教程目录
#    - docker/data/
#    - docker/.env

# 3. 提交
git commit -m "feat: 第N章简述"

# 4. 推送
git push

# 5. 打 tag 并推送
git tag ch3-数据库
git push origin ch3-数据库
```

### commit message 约定

| 前缀 | 用途 | 示例 |
|------|------|------|
| `init:` | 首次初始化 | `init: 开发环境准备` |
| `feat:` | 新功能 / 新章节 | `feat: 项目搭建` |
| `fix:` | 修复 bug | `fix: kibana 用户名` |
| `chore:` | 配置、ignore 等杂项 | `chore: 更新 gitignore` |

### tag 命名约定

与 `md/chN-xxx.md` 对齐：

```
ch1-env          ← 第 1 章 开发环境
ch2-scaffold     ← 第 2 章 项目搭建
ch3-database     ← 第 3 章（示例）
```

---

## 四、给已有 commit 补打 tag

如果之前 push 了但忘了打 tag：

```powershell
git tag ch1-env d3ba30f
git tag ch2-scaffold 4624fc7
git push origin --tags
```

---

## 五、查看某个时间步的项目

### 1. 看某次提交改了什么

```powershell
git show ch2-scaffold
git show --stat ch1-env
```

### 2. 对比两个里程碑

```powershell
git diff ch1-env ch2-scaffold
git diff ch1-env ch2-scaffold -- app/
```

### 3. 看某版本某个文件的内容（不切换工作区）

```powershell
git show ch1-env:go.work
git show ch2-scaffold:app/main.go
```

### 4. 临时切到旧版本浏览

```powershell
git checkout ch1-env
# 浏览、测试...
git checkout main    # 回到最新，继续开发
```

> 在 `checkout` 到 tag/commit 时处于 **detached HEAD**，不要在这里做 commit。纯浏览即可。

### 5. GitHub 网页查看

仓库 → **Commits** 或 **Tags** → 点击对应节点，可看该时刻全部文件快照。

---

## 六、回退

### 场景 A：已 push，想撤销但保留历史（推荐）

用 `git revert`，生成一次「反向提交」，适合已推送到 GitHub 的情况：

```powershell
git revert ch2-scaffold
git push
```

历史完整，别人仍能看到曾经做过第 2 章。

### 场景 B：本地硬回到某个 tag（慎用）

```powershell
git reset --hard ch1-env
```

本地第 2 章代码会消失。若远程也要改：

```powershell
git push --force
```

> `--force` 会改写远程历史，仅个人仓库、且确定要这么做时使用。

### 三种 reset 对比

| 命令 | 代码改动 | 暂存区 |
|------|----------|--------|
| `--soft` | 保留 | 保留 |
| `--mixed`（默认） | 保留 | 清空 |
| `--hard` | **全部丢弃** | 清空 |

---

## 七、常见问题

### 1. `git add .` 报错 elasticsearch 文件找不到

Docker 运行时 `docker/data/` 文件会变。解决：不要 `git add .`，用第三节的选择性 add；并确认 `.gitignore` 包含 `docker/data/`。

### 2. `git push` 显示 Everything up-to-date

说明没有新 commit。需要先 `git commit -m "..."` 再 push。

### 3. `git commit` 弹出 Vim / 报 empty message

用 `-m` 直接写说明：

```powershell
git commit -m "feat: 项目搭建"
```

合并时不想进编辑器：

```powershell
git pull origin main --allow-unrelated-histories --no-edit
```

### 4. PowerShell 里 curl 改 ES 密码失败

用 `Invoke-RestMethod`，不要用带 `\"` 的 curl（JSON 会被 PowerShell 吃掉）。

### 5. 教程目录会不会被误推？

根目录 `01-xxx/`、`02-xxx/` 被 `/[0-9]*/` 忽略。`md/ch2-xxx.md` 不受影响。

---

## 八、快速命令速查

```powershell
git status                          # 当前状态
git log --oneline --decorate        # 历史 + tag
git tag -l                          # 所有 tag
git diff ch1-env ch2-scaffold       # 对比两阶段
git checkout ch1-env                # 临时看旧版
git checkout main                   # 回最新
git tag ch3-xxx                     # 打 tag
git push origin ch3-xxx             # 推 tag
git push origin --tags              # 推全部 tag
```

---

## 九、推荐习惯

1. **每章 push 完立刻打 tag**，和 `md/chN-xxx.md` 同名前缀。
2. **不用 `git add .`**，用固定的 `git add .gitignore go.work app/ common/ core/ model/ md/`。
3. **教程原文永远不进仓库**，笔记写在 `md/`。
4. 回退优先 `git revert`，避免 `--force` 除非只有你一个人在用仓库。
