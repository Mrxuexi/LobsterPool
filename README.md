# LobsterPool

LobsterPool（龙虾池）是一个基于 Kubernetes 的 OpenClaw 实例开通平台。它负责接收用户填写的密钥和模板选择，自动创建对应的 K8s `Secret`、`Deployment`、`Service`，并提供实例管理界面。

它不是 OpenClaw 运行时，而是 OpenClaw 的实例 provisioner。

## 功能概览

- 用户注册、登录、修改密码
- 管理员总览、用户管理、模板管理
- 基于模板创建 OpenClaw 实例
- 自动对接 Kubernetes 或使用 Mock Provider 本地开发
- SQLite 持久化用户、模板、实例数据
- React + Ant Design 前端，Go + Gin 后端

## 技术栈

- 后端：Go 1.21、Gin、SQLite（`modernc.org/sqlite`）、client-go
- 前端：React 19、TypeScript、Vite、Ant Design 6、react-router-dom 7、i18next
- 部署：Docker、多阶段构建、Kubernetes Manifests

## 仓库结构

```text
.
├── backend/            # Go API server
│   ├── cmd/server      # 程序入口
│   └── internal/       # 配置、数据层、处理器、Provider、路由
├── frontend/           # React 前端
├── deploy/             # Dockerfile、docker-compose、K8s 清单
├── docs/               # 产品文档
└── Makefile            # 常用开发命令
```

## 核心流程

1. 用户登录后选择模板并填写 `API Key`、`Mattermost Bot Token`
2. 后端校验用户实例额度和模板信息
3. Provider 创建对应的 Kubernetes 资源
4. 实例信息写入 SQLite
5. 前端展示实例状态、访问地址和管理员统计数据

## 本地开发

### 依赖

- Go 1.21+
- Node.js 20+
- npm
- 可选：Kubernetes 集群访问权限

### 一键启动前后端

```bash
make dev
```

默认行为：

- 后端监听 `http://localhost:8080`
- 前端使用 Vite 开发服务器
- `make dev` 会自动设置 `LP_DEV_MODE=true`，因此默认使用 Mock Provider，不要求本地有 K8s 集群

### 本地轻量 K8s 自测

如果你要验证真实 Kubernetes Provider，而不是 Mock Provider，可以直接用 `kind` 起一个开发集群：

```bash
make dev-kind-up
make dev-backend-k8s
make dev-frontend
```

默认约定：

- `kind` 集群名：`lobsterpool-dev`
- kubeconfig context：`kind-lobsterpool-dev`
- LobsterPool 默认目标集群名：`kind-dev`
- 资源命名空间：`lobsterpool-local`

这套流程是轻量开发自测用，不是完整生产 K8s 部署。

### 单独启动

```bash
make dev-backend
make dev-frontend
```

### 构建

```bash
make build
make build-backend
make build-frontend
```

### 测试与检查

```bash
make test
make lint
```

端到端测试：

```bash
cd frontend && npm run test:e2e:install
make test-e2e
```

E2E 测试会自动：

- 启动一个独立的后端进程
- 使用 `LP_DEV_MODE=true` 和 Mock Provider
- 使用独立临时 SQLite 数据库
- 启动一个带 API 代理的 Vite 开发服务器

## 默认账号

数据库迁移时会自动创建默认管理员：

- 用户名：`admin`
- 密码：`admin`

首次使用后建议立即修改密码。该管理员会被强制设置为 `admin` 角色，默认不受实例数量限制。

普通用户通过注册接口创建后，默认角色为 `member`，默认最多创建 `1` 个实例。

## 配置项

后端主要通过环境变量配置：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LP_PORT` | `8080` | API 服务端口 |
| `LP_DB_PATH` | `lobsterpool.db` | SQLite 数据库路径 |
| `LP_NAMESPACE` | `lobsterpool` | 默认 Kubernetes 命名空间；多集群条目未显式指定时作为回退值 |
| `LP_KUBECONFIG` | `~/.kube/config` | 旧单集群模式或多集群条目的 kubeconfig 默认值 |
| `LP_KUBE_CLUSTERS` | 未设置 | 多集群配置，JSON 数组；每项支持 `name`、`display_name`、`namespace`、`kubeconfig`、`context`、`api_server`、`token`、`ca_file`、`insecure_skip_tls_verify` |
| `LP_DEFAULT_CLUSTER` | `default` | 默认目标集群名；创建实例时未显式指定集群则落到这里 |
| `LP_STATIC_DIR` | `./static` | 生产环境前端静态文件目录 |
| `LP_DEV_MODE` | `false` | 设为 `true` 时使用 Mock Provider |
| `LP_MOCK_PROVIDER` | 未设置 | 设为 `true` 时强制使用 Mock Provider |
| `LP_JWT_SECRET` | `lobsterpool-dev-secret-change-me` | JWT 签名密钥 |

### 多集群配置示例

通过 kube-apiserver 纳管多个集群时，推荐配置 `LP_KUBE_CLUSTERS`：

```bash
export LP_DEFAULT_CLUSTER=kind-dev
export LP_KUBE_CLUSTERS='[
  {
    "name": "kind-dev",
    "display_name": "Kind Dev",
    "namespace": "lobsterpool-local",
    "kubeconfig": "'"$HOME"'/.kube/config",
    "context": "kind-lobsterpool-dev"
  },
  {
    "name": "remote-prod",
    "display_name": "Remote Prod",
    "namespace": "lobsterpool",
    "api_server": "https://10.0.0.10:6443",
    "token": "REPLACE_ME",
    "ca_file": "/etc/lobsterpool/prod-ca.crt"
  }
]'
```

兼容性说明：

- 未设置 `LP_KUBE_CLUSTERS` 时，仍按旧单集群模式运行
- 已有实例记录会在启动时自动回填到当前 `LP_DEFAULT_CLUSTER`

## API 概览

所有接口都挂在 `/api/v1` 下。

### 健康检查

- `GET /health`

### 认证

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me`
- `POST /auth/change-password`

### 模板

- `GET /templates`
- `GET /templates/:id`

### 集群

- `GET /clusters`

### 实例

- `GET /instances`
- `POST /instances`
- `GET /instances/:id`
- `DELETE /instances/:id`

### 管理员

- `GET /admin/overview`
- `GET /admin/users`
- `PATCH /admin/users/:id/max-instances`
- `GET /admin/instances`
- `POST /admin/templates`

## Kubernetes 资源约定

- 实例资源名前缀：`claw-{instance_id}`
- 为每个实例创建：
  - `Secret`
  - `Deployment`
  - `Service`
- 默认标签：
  - `app.kubernetes.io/managed-by=lobsterpool`
  - `lobsterpool.io/instance=<instance_id>`

在真实 Kubernetes Provider 中，容器会注入以下环境变量：

- `OPENAI_API_KEY`
- `MATTERMOST_BOT_TOKEN`

## Docker

构建镜像：

```bash
make docker-build
```

镜像采用多阶段构建：

1. 构建前端静态资源
2. 构建 Go 后端二进制
3. 使用 distroless 镜像运行，并通过 `LP_STATIC_DIR=/static` 提供前端页面

## 部署文件

`deploy/` 目录包含以下资源：

- `deploy/Dockerfile`
- `deploy/docker-compose.yaml`
- `deploy/k8s/namespace.yaml`
- `deploy/k8s/rbac.yaml`
- `deploy/k8s/deployment.yaml`
- `deploy/k8s/service.yaml`
- `deploy/k8s/pvc.yaml`

## 当前默认模板

首次初始化数据库时，会自动写入一个模板：

- ID：`openclaw-mm`
- Name：`Mattermost Bot`
- Image：`registry.company.com/openclaw/mm-bot:1.0`
- Port：`8080`

## 说明

- 开发模式默认不依赖 Kubernetes，适合前后端联调
- 生产环境下，Go 服务会直接托管前端构建产物
- README 以当前仓库实现为准，若产品文档与代码不一致，请以代码行为为准
