# micro-pagode

一个面向多用户共享服务器场景的轻量级管理面板。它提供基于浏览器的文件管理、用户鉴权与权限控制，以及服务器运行状态监控与告警能力，降低频繁通过 SSH 进行日常运维的成本。

---

## 你能用它做什么

* **多用户登录与权限控制**

  * 基于 JWT 的登录鉴权，请求通过中间件统一校验 Token，并将用户身份信息写入上下文。 
  * 支持普通用户与管理员分层接口，管理员可管理用户列表、创建用户与删除用户。

* **浏览器内文件管理**

  * 支持文件/目录列表查看、创建目录、创建文件、删除、上传与下载。 
  * 支持文件搜索，便于在共享目录中快速定位目标文件。

* **轻量文件搜索快照系统**

  * 通过内存快照维护文件元数据，加速搜索而不是每次都重新遍历磁盘。
  * 使用 `RWMutex + dirty 标记 + double-check` 管理快照的并发安全读取、懒构建与失效重建。 

* **系统运行状态监控**

  * 提供 CPU、内存、磁盘、网络发送/接收等指标的查询接口。 
  * 支持监控配置管理，可设置阈值、启停开关与告警通知方式。 

* **告警与持久化**

  * 监控数据采集后统一写入存储器，并支持刷盘持久化。 
  * 当监控值超过阈值时，系统会记录告警，并支持通过 SMTP 发送邮件通知。

---

## 这个项目解决了什么问题

在多人共享服务器的使用场景里，很多日常操作本质上都很零碎：

* 想看某个目录里有哪些文件
* 想上传/下载一个文件
* 想快速搜索文件名
* 想知道服务器最近 CPU / 内存 / 磁盘 / 网络有没有异常
* 想给不同用户做最基本的权限隔离

这些需求如果全部依赖 SSH 手工处理，成本并不高到需要完整的重型运维平台，但又足够频繁，长期来看会影响效率。

`micro-pagode` 的目标就是在这个中间地带提供一个**轻量、可落地、部署简单**的方案。

---

## 未来希望增加的功能

* Web 可视化前端，完善管理界面
* 更细粒度的文件访问权限控制
* 支持更多告警通道（如 webhook / 企业微信 / Telegram）
* 更完整的监控历史数据展示与统计分析
* 支持多根目录或多实例管理

---

## 技术栈

* **Go + Gin**：提供 RESTful API 与路由管理。
* **JWT**：登录鉴权与请求身份校验。
* **GORM + MySQL**：用户信息、监控配置与告警记录持久化。
* **gopsutil**：采集 CPU / 内存 / 磁盘 / 网络状态。
* **gocron**：定时任务调度。
* **Viper**：配置管理。
* **本地文件仓库 + 快照搜索**：用于文件管理与文件名搜索。

---

## 项目结构

总体架构：**Handler → Logic → Repository / Model / Infra**

```text
internal/
├── handler/
│   ├── api/                    # HTTP 接口层
│   │   ├── userpublic.go       # 登录 / 注册
│   │   ├── userself.go         # 当前用户信息 / 修改密码 / 注销
│   │   ├── adminuser.go        # 管理员用户管理
│   │   ├── file.go             # 文件管理接口
│   │   └── system.go           # 监控与告警配置接口
│   └── result.go               # 错误处理适配
│
├── logic/
│   ├── system.go               # 系统监控、阈值检查、告警发送
│   ├── user.go                 # 用户业务逻辑
│   └── file.go                 # 文件业务逻辑
│
├── middleware/
│   ├── jwtMid.go               # JWT 鉴权中间件
│   ├── admin.go                # 管理员权限校验
│   └── logging.go              # 访问日志与 TraceID
│
├── search/
│   ├── snapshot.go             # 文件搜索快照管理
│   ├── builder.go              # 从文件系统构建快照
│   └── types.go                # 搜索元数据定义
│
├── infra/
│   ├── monitorstore/           # 监控数据存储
│   └── flusher/                # JSON 原子刷盘
│
├── repository/                 # 本地文件仓库抽象
├── model/                      # MySQL 数据模型
├── config/                     # 配置定义
└── svc/
    └── servicecontext.go       # 依赖组装与系统初始化
```
---

##  快速开始（Docker）

### 一键启动

```bash
git clone https://github.com/ZvezdyUTO/micro-pagode
cd micro-pagode
docker compose up -d
```

---

### 启动说明

* 项目内已内置默认配置（`/etc/local/api.yaml`），无需额外修改
* Docker Compose 会自动完成：

  * MySQL 服务启动
  * 后端服务构建与启动
  * 数据库连接与初始化
  * 系统默认数据创建（root 用户 / 监控配置）

---

### 服务访问

启动完成后，服务默认运行在：

```text
http://localhost:8080
```

---

## 接口示例

### 1. 注册用户

```bash
curl -s http://localhost:8080/v1/user/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"test","phone":"12345678901","password":"123456","password2":"123456"}'
```

### 2. 登录并获取 Token

```bash
curl -s http://localhost:8080/v1/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"root","password":"000000"}'
```

### 3. 查询当前用户信息

```bash
curl -s http://localhost:8080/v1/user/me \
  -H 'Authorization: Bearer <TOKEN>'
```

### 4. 查看某个目录下的文件列表

```bash
curl -s 'http://localhost:8080/v1/file?path=./data' \
  -H 'Authorization: Bearer <TOKEN>'
```

### 5. 上传文件

```bash
curl -s http://localhost:8080/v1/file/upload \
  -H 'Authorization: Bearer <TOKEN>' \
  -F 'path=./data' \
  -F 'file=@./example.txt'
```

### 6. 下载文件

```bash
curl -L 'http://localhost:8080/v1/file/download?path=./data/example.txt' \
  -H 'Authorization: Bearer <TOKEN>' \
  -o example.txt
```

### 7. 搜索文件

```bash
curl -s 'http://localhost:8080/v1/file/search?keyword=example&limit=10' \
  -H 'Authorization: Bearer <TOKEN>'
```

### 8. 查询最近一段监控数据

```bash
curl -s http://localhost:8080/v1/system/monitors \
  -H 'Authorization: Bearer <TOKEN>'
```

### 9. 更新监控告警配置

```bash
curl -s http://localhost:8080/v1/system/monitor/config \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "is_start": true,
    "cpu_limit": 90,
    "disk_limit": 80,
    "men_limit": 80,
    "net_send_limit": 1024,
    "net_recv_limit": 1024,
    "notify_type": 1,
    "email": "your_email@example.com"
  }'
```

---

## 已知说明

* 这是一个偏轻量的管理面板，不是面向大规模集群的完整运维平台。
* 当前更适合单机、多用户共享服务器场景。
* 监控通知当前主要围绕邮件告警展开。
