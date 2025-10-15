# 杭州电子科技大学食物点评系统架构设计

## 总体概述
- **目标**：提供一个供杭电学生分享校内外餐饮体验的平台，支持食物点评提交、图片上传、地址与评分信息的展示，并具备管理员审核流程。
- **技术栈**：
  - 后端：Go 1.22、Gin Web 框架、GORM ORM、SQLite（开发环境）/ 可替换成 MySQL。
  - 前端：React + Vite（单一前端框架要求）。
  - 鉴权：JWT（访问令牌 + Refresh Token 预留）、BCrypt 密码哈希。
  - 静态资源：本地 `uploads/` 目录存储图片，生产可切换为对象存储。

## 核心模块
- **用户模块**：注册、登录、角色（普通用户 / 管理员）、个人信息查询。
- **点评模块**：用户提交食物点评（名称、地址、描述、评分、图片）。点评默认进入 `pending` 状态，管理员审核后变为 `approved` 才对所有用户可见。
- **审核模块**：管理员查看待审核点评、通过或驳回；驳回时可附带备注。
- **公共浏览**：访客与普通用户可查看已审核点评，支持按评分、发布时间排序和关键字搜索（后端预留）。

## 数据模型
```mermaid
erDiagram
    USERS ||--o{ REVIEWS : "submits"
    USERS {
        uuid id
        string email
        string password_hash
        string display_name
        string role // user/admin
        datetime created_at
        datetime updated_at
    }
    REVIEWS ||--o{ REVIEW_IMAGES : "has"
    REVIEWS {
        uuid id
        string title
        string address
        text description
        float rating
        enum status // pending/approved/rejected
        string rejection_reason
        uuid author_id
        datetime created_at
        datetime updated_at
    }
    REVIEW_IMAGES {
        uuid id
        uuid review_id
        string file_path
        datetime created_at
    }
```

## API 设计（REST）
- `POST /api/v1/auth/register`：用户注册。
- `POST /api/v1/auth/login`：用户登录，返回 JWT。
- `GET /api/v1/users/me`：获取当前用户信息。
- `GET /api/v1/reviews`：列表查询，默认仅返回 `approved`，管理员可查看全部。
- `POST /api/v1/reviews`：提交点评，状态初始为 `pending`。
- `GET /api/v1/reviews/:id`：获取单条点评。
- `POST /api/v1/reviews/:id/images`：上传点评图片。
- `PUT /api/v1/reviews/:id/status`：管理员审核通过/驳回。

## 前端页面规划
- 登录 / 注册页。
- 已审核点评列表（首页），带筛选排序。
- 点评详情页（图片轮播、评分、评论内容）。
- 我的点评：显示用户提交的点评及审核状态。
- 新建点评表单。
- 管理后台：待审核列表、审核操作。

## 开发与部署
- **开发模式**：
  - 后端：`make dev` 启动 Go 服务，监听 `localhost:8080`。
  - 前端：`npm run dev` 启动 Vite 开发服务器，使用代理转发 `/api` 到 Go 服务。
- **部署建议**：
  - 使用 Docker 打包前后端；后端提供静态文件服务并暴露 REST API。
  - 生产数据库可切换为 MySQL/PostgreSQL。
  - 静态资源存储迁移到对象存储或 CDN。

## 下一步实施
1. 初始化 Go 模块，搭建项目目录结构与配置加载。
2. 实现数据库迁移、用户与点评模型、仓储层和业务服务。
3. 完成 JWT 鉴权、中间件及 REST API 路由。
4. Scaffold React + Vite 前端，配置 API 客户端与基础页面。
5. 编写 README，提供本地运行与构建指引。
