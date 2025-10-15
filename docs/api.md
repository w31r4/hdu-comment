# 后端 API 文档

本文档描述杭电食物点评系统后端（`/api/v1`）的主要 REST 接口。除特别说明外，所有请求与响应均采用 `application/json`。

## 认证

| Endpoint | Method | 说明 |
| --- | --- | --- |
| `/auth/register` | POST | 用户注册，成功后自动返回登录态 |
| `/auth/login` | POST | 用户登录获取访问/刷新令牌 |
| `/auth/refresh` | POST | 刷新访问令牌 |
| `/auth/logout` | POST | 注销（撤销刷新令牌） |

### 注册 `POST /auth/register`

请求体：

```json
{
  "email": "user@example.com",
  "password": "Password123",
  "display_name": "美食探店" 
}
```

响应：`201 Created`

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<refresh-token>",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "美食探店",
    "role": "user",
    "created_at": "2024-05-01T12:00:00Z"
  }
}
```

错误：

| 状态码 | 场景 |
| --- | --- |
| 400 | 请求参数非法 |
| 409 | 邮箱已被注册 |

### 登录 `POST /auth/login`

请求体：

```json
{
  "email": "user@example.com",
  "password": "Password123"
}
```

响应：`200 OK`，结构同注册。

错误：`401`（账号或密码错误）。

### 刷新令牌 `POST /auth/refresh`

请求体：

```json
{
  "refresh_token": "<refresh-token>"
}
```

响应：`200 OK`，返回新的访问/刷新令牌对。

错误：`401`（刷新令牌无效或过期）。

### 注销 `POST /auth/logout`

请求体：同刷新令牌。

成功：`204 No Content`。

错误：`401`（刷新令牌无效）。

## 用户

| Endpoint | Method | 说明 | 认证 |
| --- | --- | --- | --- |
| `/users/me` | GET | 获取当前登录用户信息 | 需要 `Authorization: Bearer <access_token>` |

响应：

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "display_name": "美食探店",
  "role": "user",
  "created_at": "2024-05-01T12:00:00Z"
}
```

## 点评（公共）

| Endpoint | Method | 说明 | 认证 |
| --- | --- | --- | --- |
| `/reviews` | GET | 查看已审核点评（分页、搜索、排序） | 否 |
| `/reviews/{id}` | GET | 查看点评详情。已审核点评公开，未审核/已驳回需要作者或管理员身份 | 可选 |

### 列表 `GET /reviews`

查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `page` | int，默认 1 | 页码 |
| `page_size` | int，默认 10 | 每页数量 |
| `query` | string | 按标题、地址、描述模糊搜索 |
| `sort` | `created_at` (默认) 或 `rating` | 排序字段 |
| `order` | `desc` (默认) 或 `asc` | 排序方向 |

响应：

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "学一蛋包饭",
      "address": "学一食堂二楼",
      "description": "份量足，口味偏甜",
      "rating": 4.5,
      "status": "approved",
      "images": [
        {
          "id": "uuid",
          "url": "https://..."
        }
      ],
      "created_at": "2024-05-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 42,
    "total_pages": 5
  }
}
```

### 详情 `GET /reviews/{id}`

响应格式同单条 `Review`。若点评尚未通过审核，则：

- 管理员可直接查看；
- 作者需携带有效访问令牌；
- 其他用户会收到 `403 Forbidden`。

## 点评（已登录用户）

| Endpoint | Method | 说明 | 认证 |
| --- | --- | --- | --- |
| `/reviews` | POST | 提交新的点评（初始状态为 `pending`） | 是 |
| `/reviews/me` | GET | 查看自己的点评记录（含审核状态） | 是 |
| `/reviews/{id}/images` | POST | 上传点评图片（multipart/form-data，字段名 `file`） | 是，且需作者身份 |

### 提交点评 `POST /reviews`

请求体：

```json
{
  "title": "学一蛋包饭",
  "address": "学一食堂二楼",
  "description": "份量足，口味偏甜",
  "rating": 4.5
}
```

限制：`rating` 取值 0~5。

成功：`201 Created`，返回创建后的点评（状态 `pending`）。

错误：`400`（必填字段缺失或评分越界）。

### 上传图片 `POST /reviews/{id}/images`

- Content-Type：`multipart/form-data`，字段名 `file`。
- 仅作者本人可上传。
- 成功返回 `201 Created`：

```json
{
  "id": "uuid",
  "review_id": "uuid",
  "storage_key": "...",
  "url": "https://...",
  "created_at": "2024-05-01T12:05:00Z"
}
```

## 管理员接口

管理员需在请求头中携带管理员角色的访问令牌。

| Endpoint | Method | 说明 |
| --- | --- | --- |
| `/admin/reviews/pending` | GET | 待审核点评列表（分页搜索同公共列表） |
| `/admin/reviews/{id}/approve` | PUT | 审核通过指定点评 |
| `/admin/reviews/{id}/reject` | PUT | 驳回点评并填写原因 |
| `/admin/reviews/{id}` | DELETE | 删除点评（含图片记录） |

### 审核通过 `PUT /admin/reviews/{id}/approve`

成功：`200 OK`，返回更新后的点评（状态 `approved`）。

错误：

| 状态码 | 场景 |
| --- | --- |
| 400 | 点评已被处理 |
| 404 | 点评不存在 |

### 驳回 `PUT /admin/reviews/{id}/reject`

请求体：

```json
{
  "reason": "内容重复，建议补充细节"
}
```

成功：`200 OK`，点评状态变为 `rejected` 并返回驳回原因。

### 删除点评 `DELETE /admin/reviews/{id}`

成功：`204 No Content`。后台会删除数据库记录与已存储的图片；若图片文件缺失，删除操作仍视为成功。

错误：`404`（点评不存在）。

## 错误响应格式

统一错误响应：

```json
{
  "error": "错误描述"
}
```

认证失败与鉴权失败分别返回 `401 Unauthorized`、`403 Forbidden`。服务器内部错误返回 `500 Internal Server Error`。

