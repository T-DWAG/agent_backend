# 02 - 项目搭建

> 本章目标：用 `go.work` 多模块组织代码，基于 [thunder](https://github.com/mszlu521/thunder) 基座库跑通第一个 HTTP 接口。

---

## 一、这章在干什么

上一章把 Docker 中间件和空 `go.work` 准备好了。本章做一件事：**搭好后端项目的骨架，让浏览器能访问到一个 API**。

```
go.work（工作区，统一管理多个 Go 模块）
 ├── app      ← 主服务：HTTP 入口、路由、业务 Handler
 ├── common   ← 公共依赖（thunder 在这里 go get，方便统一维护）
 ├── core     ← 核心工具（AI / Eino，后续章节再用）
 └── model    ← 数据模型（后续章节再用）
```

### 一次请求的完整链路

```
浏览器  GET /api/v1/auth/register
   │
   ▼
main.go          启动程序，加载配置、日志、Server
   │
   ▼
inits.Init()     把所有 Router 注册到 Server
   │
   ▼
router/auth.go   声明 URL → Handler 的映射
   │
   ▼
auths/handler.go  处理业务，调用 res.Success 返回 JSON
```

### thunder 是什么

老师封装的**基础组件库**，本章用到：

| 包 | 作用 |
|---|---|
| `config` | 读取 `etc/config.yml` |
| `logs` | 初始化日志 |
| `server` | 封装 Gin HTTP Server |
| `res` | 统一 JSON 响应格式 |

---

## 二、架构特点（教程原文 + 理解）

| 特点 | 含义 |
|---|---|
| 模块化 | app / common / core / model 各司其职，后期可独立演进 |
| 分层 | router（表现层）→ handler/service（业务层）→ repository（数据层，后面加） |
| 微服务友好 | 模块边界清晰，将来可拆成独立服务 |
| AI 原生 | core 模块预留 Eino 集成（本章不涉及） |

---

## 三、最终目录结构

```
backend/
├── go.work
├── md/                          ← 章节学习笔记（本章文档）
├── app/
│   ├── main.go                  ← 入口，必须在 app 根目录，不能放 internal
│   ├── go.mod
│   ├── etc/
│   │   └── config.yml           ← 配置文件
│   └── internal/                ← 内部代码，外部模块不可 import
│       ├── inits/
│       │   └── init.go          ← 初始化：注册路由、中间件等
│       ├── router/
│       │   ├── auth.go          ← 认证相关路由
│       │   └── event.go         ← 事件路由（占位，后续扩展）
│       └── auths/
│           └── handler.go       ← 认证 Handler
├── common/
│   └── go.mod                   ← thunder 依赖在这里 go get
├── core/
│   └── go.mod                   ← 空模块，后续使用
└── model/
    └── go.mod                   ← 空模块，后续使用
```

---

## 四、逐步搭建（命令 + 逻辑）

### Step 1：创建 4 个子模块

```powershell
cd E:\data\code\eino\project\backend

mkdir common, core, model, app

cd common
go mod init common
go get github.com/mszlu521/thunder
cd ..

cd core && go mod init core && cd ..
cd model && go mod init model && cd ..
cd app && go mod init app && cd ..
```

**逻辑：**

- `go mod init <名>` 给每个目录生成独立的 `go.mod`，各自是一个 Go 模块。
- thunder 放在 **common** 里 `go get`，是教程约定的「依赖统一管理」方式；app 代码里仍会直接 import thunder，需要在 app 目录再 `go mod tidy` 拉依赖。
- core / model 这章只建空壳，避免后面再加模块时改 go.work。

---

### Step 2：写入 go.work

```powershell
go work use ./app ./common ./core ./model
```

**逻辑：**

- `go.work` 把 4 个模块绑成一个**工作区**，本地开发时模块之间可以直接引用，不用发到 GitHub。
- 打开 Goland / VS Code 时，以 `backend/` 为根目录打开，IDE 才能识别整个 workspace。

生成结果：

```go
go 1.26.2

use (
    ./app
    ./common
    ./core
    ./model
)
```

---

### Step 3：创建 app 子目录

```powershell
mkdir app\etc
mkdir app\internal\inits
mkdir app\internal\router
mkdir app\internal\auths
```

**逻辑：**

- `etc/` 放配置文件，thunder 默认读 `etc/config.yml`（相对**运行目录**）。
- `internal/` 是 Go 语言约定：该目录下的包**只能被 app 模块内部 import**，防止外部误用内部实现。

---

### Step 4：编写 `app/etc/config.yml`

```yaml
app:
  name: "chuanxing-ai"    # 应用标识名，可自定义，不影响端口和路由
server:
  mode: "debug"           # Gin 模式：debug / release / test
  host: "127.0.0.1"
  port: 8888
  readTimeout: 10s
  writeTimeout: 10s
  cors:
    - "*"
log:
  level: "info"
  format: "pretty"        # 开发用 pretty，生产建议 json
  addSource: true
```

**逻辑：**

| 配置块 | 作用 |
|---|---|
| `app.name` | 服务名称标识，日志/监控可能用到，**可随意改** |
| `server.*` | 监听地址、端口、超时、CORS |
| `log.*` | 日志级别和格式 |

> 教程原文写 `mszlu-ai`，那是作者项目名。本项目改为 `chuanxing-ai`（或 `agent-backend` 等均可）。

---

### Step 5：编写 `app/main.go`（入口）

```go
package main

import (
    "app/internal/inits"

    "github.com/mszlu521/thunder/config"
    "github.com/mszlu521/thunder/logs"
    "github.com/mszlu521/thunder/server"
)

func main() {
    config.Init()              // 1. 加载 etc/config.yml
    conf := config.GetConfig()
    logs.Init(conf.Log)        // 2. 初始化日志
    s := server.NewServer(conf) // 3. 创建 Gin Server
    inits.Init(s, conf)        // 4. 注册路由等业务初始化
    s.Start()                  // 5. 阻塞启动，监听端口
}
```

**逻辑：**

启动顺序固定：**配置 → 日志 → Server → 业务初始化 → 监听**。

`main.go` 只做编排，不写业务，保持入口简洁。

> ⚠️ **常见错误**：把 `main.go` 放到 `app/internal/` 下，运行 `go run main.go` 会报 `The system cannot find the file specified`。入口必须在 `app/main.go`。

---

### Step 6：编写 `app/internal/inits/init.go`

```go
package inits

import (
    "app/internal/router"

    "github.com/mszlu521/thunder/config"
    "github.com/mszlu521/thunder/server"
)

func Init(s *server.Server, conf *config.Config) {
    s.RegisterRouters(&router.Event{}, &router.AuthRouter{})
}
```

**逻辑：**

- `inits` 是**初始化中心**：路由、中间件、数据库连接等以后都从这里统一注册。
- `RegisterRouters` 接收实现了路由接口的结构体，thunder 内部会调用它们的 `Register` 方法。
- 新增功能模块时，在这里加一行 `&router.XxxRouter{}` 即可。

---

### Step 7：编写 Handler — `app/internal/auths/handler.go`

```go
package auths

import (
    "github.com/gin-gonic/gin"
    "github.com/mszlu521/thunder/res"
)

type Handler struct{}

func NewHandler() *Handler {
    return &Handler{}
}

func (h *Handler) Register(c *gin.Context) {
    res.Success(c, nil)
}
```

**逻辑：**

- **Handler** = 处理 HTTP 请求的代码，对应「表现层 / 控制器」。
- `auths` 包代表「认证模块」，后面 login、logout 等 Handler 都放这里。
- `res.Success(c, nil)` 走 thunder 统一响应格式，前端拿到的 JSON 结构一致。

---

### Step 8：编写路由 — `app/internal/router/auth.go`

```go
package router

import (
    "app/internal/auths"

    "github.com/gin-gonic/gin"
)

type AuthRouter struct{}

func (u *AuthRouter) Register(engine *gin.Engine) {
    userGroup := engine.Group("/api/v1/auth")
    {
        userHandler := auths.NewHandler()
        userGroup.GET("/register", userHandler.Register)
    }
}
```

**逻辑：**

- **Router** 只负责「URL → Handler 方法」的映射，不写业务逻辑。
- `engine.Group("/api/v1/auth")` 创建路由组，该组下所有接口前缀都是 `/api/v1/auth`。
- 分层好处：改 URL 只动 router；改业务只动 handler。

---

### Step 9：事件路由占位 — `app/internal/router/event.go`

```go
package router

type Event struct{}

func (*Event) Register() {
    // TODO: 注册事件相关路由
}
```

**逻辑：**

- 占位结构体，满足 `RegisterRouters` 的参数要求。
- 后续章节会在这里注册事件总线、WebSocket 等。

---

### Step 10：拉依赖

```powershell
cd E:\data\code\eino\project\backend\app
go mod tidy
```

**逻辑：**

- 根据 import 自动写入 `go.mod` / `go.sum`。
- app 直接 import 了 thunder，即使 common 里已经 go get 过，app 模块仍需要自己的 require 记录。

---

## 五、运行与验证

```powershell
cd E:\data\code\eino\project\backend\app
go run main.go
```

浏览器或 curl 访问：

```
http://127.0.0.1:8888/api/v1/auth/register
```

PowerShell：

```powershell
curl.exe http://127.0.0.1:8888/api/v1/auth/register
```

**期望：** 返回 JSON 成功响应（如 `code: 0`）。

### 启动注意

| 项 | 说明 |
|---|---|
| 工作目录 | 必须在 `app/` 下运行，`config.yml` 路径才正确 |
| Goland | Run Configuration 的 Working directory 设为 `$ProjectFileDir$/app` |
| 端口占用 | 8888 被占用时改 `config.yml` 的 `server.port` |

---

## 六、各层职责速查

```
main.go          程序入口，编排启动流程
    ↓
inits/           初始化中心（路由、DB、中间件…）
    ↓
router/          URL 映射（表现层）
    ↓
auths/handler    业务处理（本章只有空实现）
    ↓
res.Success      统一响应（thunder 提供）
```

---

## 七、常见问题

### 1. `main.go: The system cannot find the file specified`

`main.go` 误放在 `app/internal/` 下。移到 `app/main.go`。

### 2. 配置读不到 / 端口不对

没有从 `app/` 目录启动。thunder 默认找 `./etc/config.yml`，相对的是**进程工作目录**。

### 3. `app.name` 要改吗？

随意。只是应用标识，不影响功能。教程 `mszlu-ai` → 本项目 `chuanxing-ai`。

### 4. common / core / model 这章为什么是空的？

分模块是**先搭架子**，后面章节往里面填：common 放工具、model 放 GORM 模型、core 放 Eino AI 逻辑。

---

## 八、本章完成标准

- [x] 4 个模块 + `go.work` 配置完成
- [x] `go run main.go` 无报错启动
- [x] `GET /api/v1/auth/register` 返回正常 JSON
- [ ] （可选）提交到 GitHub：`git commit -m "feat: 项目骨架搭建"`

---

## 九、下一章预告

- 接入 PostgreSQL / Redis（Docker 已就绪）
- model 模块定义数据实体
- 完善 auths 注册/登录真实业务逻辑
