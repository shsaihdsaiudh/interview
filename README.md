# 面试互助平台

校内模拟面试配对平台 —— 找到志同道合的同学，进行一对一模拟面试练习。

## 技术栈

| 层 | 技术 |
|---|------|
| **前端** | React 19 + TypeScript + Vite + Tailwind CSS v4 |
| **后端** | Go 1.26 + Gin（六边形架构） |
| **数据库** | PostgreSQL 17 |
| **认证** | JWT + bcrypt |

## 快速开始

### 1. 启动数据库

```bash
docker compose up -d
```

PostgreSQL 会在 `localhost:5432` 启动，初始化脚本自动建表。

### 2. 启动后端

```bash
cd server
go run cmd/main.go
```

服务运行在 `http://localhost:8080`。

### 3. 启动前端

```bash
cd web
npm install
npm run dev
```

打开 `http://localhost:5173`。

> 注册时需要使用 `@std.uestc.edu.cn` 结尾的邮箱。开发阶段验证码会打印在后端控制台。

### 4. 注入测试数据（可选）

```bash
cd server
go run ./cmd/seed/
```

向数据库注入 5 个测试用户（密码统一为 `test123`）、招募卡片、未来 7 天随机空闲时间段、以及 3 条示例预约记录。**幂等**：重复运行不会产生重复数据。

| 用户 | 邮箱 | 角色 | 技能 |
|------|------|------|------|
| 张三 | zhangsan@test.com | both | React, TypeScript, Node.js |
| 李四 | lisi@test.com | interviewer | Go, Python, Docker |
| 王五 | wangwu@test.com | interviewee | Java, Spring, MySQL |
| 赵六 | zhaoliu@test.com | both | Vue, JavaScript, CSS |
| 钱七 | qianqi@test.com | interviewer | Rust, C++, 系统设计 |

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 后端监听端口 |
| `DATABASE_URL` | `postgres://interview:interview123@localhost:5432/interview_platform?sslmode=disable` | 数据库连接 |

## 项目结构

```
├── docker-compose.yml          # PostgreSQL 17
├── docker/init.sql             # 建表 DDL
├── server/                     # Go 后端
│   ├── cmd/
│   │   ├── main.go              # 入口
│   │   └── seed/main.go         # 测试数据注入脚本
│   ├── domain/                 # 领域层（entity + repository 接口）
│   ├── application/            # 应用服务层
│   ├── infrastructure/         # 基础设施（PostgreSQL、JWT）
│   └── interfaces/             # HTTP 层（Gin handler + 路由 + 中间件）
└── web/                        # React 前端
    └── src/
        ├── components/         # 通用组件（Navbar、Icons）
        ├── pages/              # 页面组件
        ├── api/client.ts       # axios 封装
        └── router/             # 路由配置
```

## 功能

- **用户注册/登录** — `@std.uestc.edu.cn` 邮箱验证码注册 + JWT 认证
- **用户广场** — 浏览所有已验证用户，按面试方向标签筛选
- **用户详情** — 查看个人资料 + 空闲时间段 → 选择时间发起预约
- **预约管理** — 双 Tab（收到/发出），接受或拒绝预约请求
- **个人设置** — 编辑资料（昵称/院系/标签/联系方式）+ 管理空闲时间段

## API 概览

所有接口前缀 `/api/v1`。

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/ping` | 健康检查 |
| `POST` | `/auth/send-code` | 发送注册邮箱验证码 |
| `POST` | `/auth/register` | 使用验证码注册（限制 `@std.uestc.edu.cn` 邮箱） |
| `POST` | `/auth/login` | 登录 → 返回 JWT |
| `GET` | `/users` | 用户列表 |
| `GET` | `/users/:id` | 用户详情 + 空闲时间；登录且双方有已接受预约时返回联系方式 |

### 需认证（Header: `Authorization: Bearer <token>`）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/auth/me` | 当前用户信息 |
| `GET` | `/profile` | 个人完整资料 |
| `PUT` | `/profile` | 更新个人资料 |
| `GET` | `/availability` | 我的空闲时间 |
| `POST` | `/availability` | 添加空闲时间 |
| `DELETE` | `/availability/:id` | 删除空闲时间 |
| `POST` | `/appointments` | 发起预约 |
| `GET` | `/appointments?role=mentor\|student` | 查看预约 |
| `PUT` | `/appointments/:id/accept` | 接受预约 |
| `PUT` | `/appointments/:id/reject` | 拒绝预约 |

## 架构

后端采用**六边形架构（Ports & Adapters）**：

- **Domain** — 纯业务实体 + Repository 接口，不依赖任何框架
- **Application** — 编排领域逻辑，注入接口
- **Infrastructure** — PostgreSQL 实现（pgx 裸 SQL）、JWT 工具
- **Interfaces** — Gin HTTP handler + 中间件，依赖注入在 `main.go` 组装

前端采用**纯 SPA**，`react-router-dom` 客户端路由，`axios` + 拦截器自动附带 JWT。
