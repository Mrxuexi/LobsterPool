# 龙虾池（LobsterPool）PRD v0.1

## 一、产品定义

**龙虾池是一个基于 Kubernetes 的 OpenClaw 实例快速开通平台。**

目标：

> 让内部同事通过简单表单即可获取一个可运行的 OpenClaw 实例。

用户无需了解：

* Kubernetes
* Deployment
* Secret
* Service
* Pod

平台会自动完成所有部署动作。

---

# 二、系统角色

系统中有两类角色：

## 1 用户（使用者）

内部同事。

只需要：

* 填写 API Key
* 填写 Mattermost Bot Token
* 点击创建实例

即可获得 OpenClaw。

---

## 2 开发者（Claw Provider）

开发者负责提供：

* OpenClaw 镜像
* 默认配置模板

例如：

```
registry.company.com/openclaw/mm-bot:1.0
```

---

# 三、系统架构

架构非常简单：

```
用户
   │
   ▼
龙虾池 Web UI
   │
   ▼
龙虾池 API Server
   │
   ▼
Kubernetes Provider
   │
   ▼
Kubernetes Cluster
```

龙虾池本身 **不运行 OpenClaw**。

它只是：

> Kubernetes 应用创建器。

---

# 四、核心功能

## 1 创建 OpenClaw 实例

用户填写：

```
API Key
Mattermost Bot Token
```

系统执行：

1 创建 K8s Secret
2 创建 Deployment
3 创建 Service
4 返回访问地址

---

## 2 查看实例

用户可以查看：

```
实例名称
状态
访问地址
创建时间
```

状态来自：

```
Kubernetes Pod Status
```

例如：

```
Pending
Starting
Running
Failed
```

---

## 3 删除实例

删除时系统会清理：

```
Deployment
Service
Secret
```

---

# 五、镜像管理模型

镜像由开发者提供。

龙虾池只维护一个 **Claw Template 列表**。

示例：

```
ClawTemplate
```

字段：

```
id
name
description
image
version
default_port
created_at
```

示例：

```
id: openclaw-mm
name: Mattermost Bot
image: registry.company.com/openclaw/mm-bot:1.0
port: 8080
```

用户创建实例时：

选择模板即可。

---

# 六、实例数据模型

```
Instance
```

字段：

```
id
name
template_id
user_id
namespace
deployment_name
service_name
status
endpoint
created_at
```

---

# 七、Kubernetes 实现设计

创建实例时平台生成以下资源。

---

## 1 Secret

用于保存敏感信息。

```
apiVersion: v1
kind: Secret
```

内容：

```
OPENAI_API_KEY
MATTERMOST_BOT_TOKEN
```

---

## 2 Deployment

```
apiVersion: apps/v1
kind: Deployment
```

核心字段：

```
image: registry.company.com/openclaw/mm-bot:1.0
```

环境变量来自：

```
secretRef
```

---

示例：

```
env:
- name: OPENAI_API_KEY
  valueFrom:
    secretKeyRef:
      name: claw-secret
      key: api_key

- name: MATTERMOST_BOT_TOKEN
  valueFrom:
    secretKeyRef:
      name: claw-secret
      key: mm_token
```

---

## 3 Service

```
kind: Service
type: ClusterIP
```

用于内部访问。

如果需要外网访问：

可以额外创建：

```
Ingress
```

---

# 八、Kubernetes Provider 接口

虽然目前只支持 K8s，但仍建议抽象一层。

接口：

```
createInstance(input)
deleteInstance(instanceId)
getInstanceStatus(instanceId)
getInstanceEndpoint(instanceId)
```

未来如果需要：

* Serverless
* VM

可以扩展。

---

# 九、实例命名规则

建议统一命名：

```
claw-{instance_id}
```

示例：

```
deployment: claw-123
service: claw-123
secret: claw-123
```

namespace：

```
lobsterpool
```

---

# 十、用户流程

用户流程非常简单：

### Step 1

进入龙虾池。

---

### Step 2

点击：

```
创建实例
```

---

### Step 3

填写：

```
API Key
Mattermost Bot Token
```

---

### Step 4

点击：

```
启动实例
```

---

### Step 5

获得：

```
实例地址
```

---

# 十一、安全设计

由于涉及 API Key，需要注意：

### 1 Token 不写入日志

日志禁止打印：

```
API Key
Bot Token
```

---

### 2 Secret 存储在 Kubernetes

平台不长期保存 token。

---

### 3 页面只显示

```
已配置
未配置
```

---

# 十二、登录系统（预留）

当前阶段可以：

* 内网访问
* 不强制登录

但预留接口：

```
/auth/login
/auth/logout
/auth/me
/auth/oauth/github
```

未来可以接入：

* GitHub
* 企业 SSO
* OIDC

---

# 十三、MVP范围

v0.1 需要实现：

### Web

* 实例列表
* 创建实例
* 实例详情
* 删除实例

---

### Backend

* Instance API
* Kubernetes Provider
* Template 管理

---

### Kubernetes

自动创建：

```
Secret
Deployment
Service
```

---

# 十四、成功标准

龙虾池 MVP 成功的标准：

一个 **从未部署过 OpenClaw 的同事**：

5分钟内完成：

```
填写参数
点击创建
获得实例
```

---

# 十五、一句话总结

龙虾池不是 OpenClaw 的分发平台，而是：

> **一个 Kubernetes 上的 OpenClaw 实例开通器。**
