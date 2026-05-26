# API 接口设计

> 文档版本: v2.0 | 最后更新: 2026-05-26
> 基础路径: `/api/v1`
> 数据格式: JSON

---

## 1. 通用规范

### 1.1 统一响应格式

**成功响应**:
```json
{
  "code": 200,
  "message": "ok",
  "data": { ... }
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "invalid input: amount must be positive"
}
```

### 1.2 HTTP 状态码使用

| 状态码 | 含义 | 场景 |
|--------|------|------|
| 200 | 成功 | GET, PUT, DELETE |
| 201 | 创建成功 | POST |
| 400 | 请求参数错误 | 参数校验失败 |
| 401 | 未认证 | Token 缺失/过期/已撤销 |
| 404 | 资源不存在 | 查询 ID 不存在 |
| 409 | 资源冲突 | 用户名/邮箱已存在, 分类有交易引用 |
| 500 | 服务器内部错误 | 数据库异常等 |

### 1.3 认证方式

所有受保护接口 (除 health, auth/register, auth/login 外) 需要在请求头中携带:

```
Authorization: Bearer <jwt_token>
```

Token 通过 `POST /auth/login` 或 `POST /auth/register` 获取。

### 1.4 分页规范

分页参数统一为 `page` 和 `page_size`, 响应格式:

```json
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

- 默认 `page=1`, `page_size=20`
- `page_size` 范围: 1–100, 超出自动修正为 20

---

## 2. 接口清单

### 2.1 健康检查

#### `GET /api/v1/health`

无需认证。返回系统健康状态。

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "status": "ok",
    "db": "ok"
  },
  "message": "ok"
}
```

---

### 2.2 认证 (Auth)

#### `POST /api/v1/auth/register`

注册新用户。成功后自动创建默认账本和分类, 返回 JWT Token。

**请求体**:
```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "secret123"
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| username | string | 是 | 2–50 字符, 唯一 |
| email | string | 是 | 有效邮箱, 唯一 |
| password | string | 是 | 6–100 字符 |

**响应** (201):
```json
{
  "code": 201,
  "message": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "alice",
      "email": "alice@example.com",
      "is_active": true,
      "created_at": "2026-05-26T10:00:00Z"
    }
  }
}
```

#### `POST /api/v1/auth/login`

用用户名和密码登录, 返回 JWT Token。

**请求体**:
```json
{
  "username": "alice",
  "password": "secret123"
}
```

**响应** (200): 同上, 含 token + user。

#### `GET /api/v1/auth/me`

获取当前登录用户信息。需认证。

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "alice",
    "email": "alice@example.com",
    "is_active": true,
    "created_at": "2026-05-26T10:00:00Z"
  }
}
```

#### `POST /api/v1/auth/logout`

登出当前用户, 将当前 Token 加入黑名单。需认证。

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "message": "logged out"
  }
}
```

#### `PUT /api/v1/auth/password`

修改密码。需认证。

**请求体**:
```json
{
  "old_password": "secret123",
  "new_password": "newSecret456"
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| old_password | string | 是 | 当前密码, 用于验证身份 |
| new_password | string | 是 | 6-100 字符, 新密码 |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "message": "password updated"
  }
}
```

**错误码**:

| 状态码 | message | 场景 |
|--------|---------|------|
| 400 | `"old password is incorrect"` | 旧密码不匹配 |
| 400 | `"new password must be at least 6 characters"` | 新密码过短 |

#### `PUT /api/v1/auth/email`

修改邮箱。需认证。

**请求体**:
```json
{
  "email": "newemail@example.com"
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| email | string | 是 | 有效邮箱格式, 唯一 |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "message": "email updated"
  }
}
```

**错误码**:

| 状态码 | message | 场景 |
|--------|---------|------|
| 400 | `"invalid email format"` | 邮箱格式不正确 |
| 409 | `"email already in use"` | 邮箱已被其他用户使用 |

---

### 2.3 账本 (Ledger)

所有接口需认证。

#### `GET /api/v1/ledgers`

获取当前用户所有账本列表。

**查询参数**: 无

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "user_id": "...",
      "name": "日常账本",
      "description": "日常收支记录",
      "base_currency": "CNY",
      "icon": null,
      "color": null,
      "is_archived": false,
      "sort_order": 0,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

#### `POST /api/v1/ledgers`

创建新账本。

**请求体**:
```json
{
  "name": "旅行账本",
  "description": "旅行期间开销",
  "base_currency": "CNY",
  "icon": "✈️",
  "color": "#1890ff"
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| name | string | 是 | 最多 100 字符 |
| description | string | 否 | 最多 500 字符 |
| base_currency | string | 否 | 默认 "CNY" |
| icon | string | 否 | 建议 Emoji |
| color | string | 否 | 十六进制颜色 |

**响应** (201): 返回创建的账本对象。

#### `GET /api/v1/ledgers/:ledger_id`

获取单个账本详情。需验证归属权。

**响应** (200): 返回账本对象。

#### `PUT /api/v1/ledgers/:ledger_id`

更新账本信息。支持部分更新。

**请求体** (所有字段可选):
```json
{
  "name": "新账本名",
  "description": "新描述",
  "base_currency": "USD",
  "icon": "🏠",
  "color": "#52c41a",
  "is_archived": false,
  "sort_order": 1
}
```

**响应** (200): 返回更新后的账本对象。

#### `DELETE /api/v1/ledgers/:ledger_id`

删除账本及其关联的交易和分类 (级联删除)。

**响应** (200):
```json
{
  "code": 200,
  "message": "ok"
}
```

#### `GET /api/v1/ledgers/:ledger_id/summary`

获取账本汇总统计。

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "total_income": 15000.00,
    "total_expense": 8210.50,
    "balance": 6789.50,
    "base_currency": "CNY",
    "expense_by_category": [
      {
        "category_id": "550e8400-...",
        "category_name": "餐饮",
        "category_icon": "🍽️",
        "total": 3200.00,
        "count": 45
      },
      {
        "category_id": "...",
        "category_name": "交通",
        "category_icon": "🚗",
        "total": 580.00,
        "count": 12
      }
    ]
  }
}
```

#### `GET /api/v1/ledgers/:ledger_id/export`

导出交易记录。需认证。

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| format | string | 否 | 导出格式: `csv` 或 `json`, 默认 `csv` |
| start_date | string | 否 | 开始日期 `"2006-01-02"` |
| end_date | string | 否 | 结束日期 `"2006-01-02"` |

**响应** (200): 返回文件内容, Content-Type 根据 format 自动设置。

- `format=csv` → `text/csv`, 文件名 `{账本名}_交易记录_{日期}.csv`
- `format=json` → `application/json`, 文件名 `{账本名}_交易记录_{日期}.json`

**错误码**:

| 状态码 | message | 场景 |
|--------|---------|------|
| 400 | `"invalid format, must be csv or json"` | format 参数不合法 |

#### `GET /api/v1/ledgers/:ledger_id/tags`

获取指定账本所有交易标签列表。需认证。

**查询参数**: 无

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": ["午餐", "外卖", "Freelance", "通勤"]
}
```

---

### 2.4 分类 (Category)

所有接口需认证。

#### `GET /api/v1/ledgers/:ledger_id/categories`

获取指定账本的分类列表 (树形结构, 根节点带 children 数组)。

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "id": "...",
      "user_id": "...",
      "ledger_id": null,
      "name": "餐饮",
      "type": "expense",
      "icon": "🍽️",
      "color": null,
      "parent_id": null,
      "sort_order": 0,
      "is_active": true,
      "children": [
        {
          "id": "...",
          "name": "外卖",
          "type": "expense",
          "icon": "📦",
          "parent_id": "...",
          "children": []
        }
      ]
    }
  ]
}
```

此接口有缓存, TTL 10 分钟。增删改分类时自动失效。

#### `POST /api/v1/categories`

创建新分类。

**请求体**:
```json
{
  "name": "外卖",
  "type": "expense",
  "icon": "📦",
  "color": "#ff4d4f",
  "ledger_id": "550e8400-...",
  "parent_id": "550e8400-..."
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| name | string | 是 | 最多 50 字符 |
| type | string | 是 | "income" 或 "expense" |
| icon | string | 否 | Emoji |
| color | string | 否 | 十六进制 |
| ledger_id | string | 否 | null=全局分类 |
| parent_id | string | 否 | 父分类 ID |

**响应** (201): 返回创建的分类对象。

#### `PUT /api/v1/categories/:id`

更新分类。支持部分更新。

#### `DELETE /api/v1/categories/:id`

删除分类。如果该分类下有关联的交易记录, 返回 409 禁止删除, 建议改为停用 (is_active=false)。

---

### 2.5 交易记录 (Transaction)

所有接口需认证。

#### `GET /api/v1/ledgers/:ledger_id/transactions`

获取交易记录列表, 支持多维筛选和分页。

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码, 默认 1 |
| page_size | int | 否 | 每页条数, 默认 20, 最大 100 |
| type | string | 否 | 筛选: "income" 或 "expense" |
| category_id | string | 否 | 按分类筛选 |
| start_date | string | 否 | 开始日期 "2006-01-02" |
| end_date | string | 否 | 结束日期 "2006-01-02" |
| keyword | string | 否 | 描述关键词搜索 (ILIKE) |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-...",
        "ledger_id": "...",
        "user_id": "...",
        "category_id": "...",
        "type": "expense",
        "amount": 35.00,
        "currency": "CNY",
        "exchange_rate": 1.0,
        "base_amount": 35.00,
        "description": "午餐",
        "transaction_date": "2026-01-15",
        "tags": "午餐,外卖",
        "is_reconciled": false,
        "created_at": "2026-01-15T12:00:00Z",
        "updated_at": "2026-01-15T12:00:00Z",
        "category": {
          "id": "...",
          "name": "餐饮",
          "icon": "🍽️",
          "type": "expense"
        }
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

#### `POST /api/v1/transactions`

创建交易记录。自动处理多币种折算。

**请求体**:
```json
{
  "ledger_id": "550e8400-...",
  "category_id": "550e8400-...",
  "type": "expense",
  "amount": 35.00,
  "currency": "CNY",
  "description": "午餐",
  "transaction_date": "2026-01-15",
  "tags": ["午餐", "外卖"]
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| ledger_id | string | 是 | UUID |
| category_id | string | 是 | UUID |
| type | string | 是 | "income" 或 "expense" |
| amount | number | 是 | 大于 0 |
| currency | string | 否 | 默认 "CNY" |
| description | string | 否 | 文本 |
| transaction_date | string | 否 | 默认当天 |
| tags | string[] | 否 | 标签数组 |

**自动折算逻辑**:
- 如果 `currency == ledger.base_currency`: `exchange_rate=1.0`, `base_amount=amount`
- 如果 `currency ≠ ledger.base_currency`: 查询 `ExchangeRate` 表取最近汇率, 计算 `base_amount = amount * rate`
- 如果查询不到汇率: 以 `rate=1.0` 处理, 即原金额直接作为本位币金额

#### `PUT /api/v1/transactions/:id`

更新交易。支持部分更新。金额/币种变更时重新计算 base_amount。

**请求体** (所有字段可选):
```json
{
  "category_id": "new-category-id",
  "type": "income",
  "amount": 100.00,
  "currency": "USD",
  "description": "自由职业收入",
  "transaction_date": "2026-03-01",
  "tags": ["Freelance"],
  "is_reconciled": true
}
```

#### `DELETE /api/v1/transactions/:id`

删除交易记录。

#### `POST /api/v1/transactions/batch-delete`

批量删除交易记录。需认证。

**请求体**:
```json
{
  "ids": ["id-1", "id-2", "id-3"]
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| ids | string[] | 是 | 至少 1 个, 最多 100 个 UUID |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "deleted_count": 3
  }
}
```

**错误码**:

| 状态码 | message | 场景 |
|--------|---------|------|
| 400 | `"ids must not be empty"` | 未提供 ID 列表 |
| 400 | `"too many ids, max 100"` | 超过批量上限 |

#### `PUT /api/v1/transactions/batch-update`

批量修改交易分类。需认证。

**请求体**:
```json
{
  "ids": ["id-1", "id-2"],
  "category_id": "new-category-id"
}
```

| 字段 | 类型 | 必填 | 约束 |
|------|------|------|------|
| ids | string[] | 是 | 至少 1 个, 最多 100 个 UUID |
| category_id | string | 是 | 目标分类 UUID |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "updated_count": 2
  }
}
```

**错误码**:

| 状态码 | message | 场景 |
|--------|---------|------|
| 400 | `"ids must not be empty"` | 未提供 ID 列表 |
| 404 | `"category not found"` | 目标分类不存在 |
| 409 | `"category does not belong to this ledger"` | 分类不属于该账本 |

---

### 2.6 汇率 (Exchange Rate)

所有接口需认证。

#### `GET /api/v1/exchange-rates`

获取汇率列表。支持筛选。

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| date | string | 否 | 按日期筛选 |
| from | string | 否 | 源币种 |
| to | string | 否 | 目标币种 |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "id": "...",
      "from_currency": "USD",
      "to_currency": "CNY",
      "rate": 7.24000000,
      "date": "2026-05-26",
      "source": "manual",
      "created_at": "2026-05-26T10:00:00Z"
    }
  ]
}
```

#### `POST /api/v1/exchange-rates`

创建汇率。同日期 + 同币种对自动覆盖 (upsert 语义)。

**请求体**:
```json
{
  "from_currency": "USD",
  "to_currency": "CNY",
  "rate": 7.24,
  "date": "2026-05-26",
  "source": "bank-of-china"
}
```

#### `GET /api/v1/exchange-rates/latest`

获取每种币对的最新一条汇率。

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "from_currency": "USD",
      "to_currency": "CNY",
      "rate": 7.24000000,
      "date": "2026-05-26"
    }
  ]
}
```

#### `DELETE /api/v1/exchange-rates/:id`

删除汇率。

---

### 2.7 分析 (Analytics)

所有接口需认证。

#### `GET /api/v1/ledgers/:ledger_id/monthly-trend`

获取月度收支趋势。用于折线图展示。

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| months | int | 否 | 最近月数, 默认 6, 最大 24 |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "month": "2026-01",
      "income": 15000.00,
      "expense": 8210.50
    },
    {
      "month": "2026-02",
      "income": 12000.00,
      "expense": 7350.00
    }
  ]
}
```

- 按月聚合 `transaction_date`, 分别计算收入和支出合计
- 使用 `base_amount` 以确保多币种环境下数据可比

#### `GET /api/v1/ledgers/:ledger_id/category-breakdown`

获取分类支出/收入分布。用于环形图展示。

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 否 | 开始日期 `"2006-01-02"` |
| end_date | string | 否 | 结束日期 `"2006-01-02"` |
| type | string | 否 | 筛选: `"income"` 或 `"expense"`, 默认 `"expense"` |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "category_id": "...",
      "category_name": "餐饮",
      "category_icon": "🍽️",
      "total": 3200.00,
      "percentage": 38.97,
      "count": 45
    },
    {
      "category_id": "...",
      "category_name": "交通",
      "category_icon": "🚗",
      "total": 580.00,
      "percentage": 7.06,
      "count": 12
    }
  ]
}
```

- `total` 为该分类金额合计 (base_amount)
- `percentage` 为该分类占总金额的百分比 (四舍五入保留 2 位小数)
- 前 8 大分类 + "其他" 合并项

#### `GET /api/v1/ledgers/:ledger_id/daily-transactions`

获取日历每日汇总。用于日历视图。

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| year | int | 是 | 年份, 如 2026 |
| month | int | 是 | 月份, 1-12 |

**响应** (200):
```json
{
  "code": 200,
  "message": "ok",
  "data": [
    {
      "date": "2026-01-01",
      "income": 0.00,
      "expense": 235.50,
      "count": 3
    },
    {
      "date": "2026-01-02",
      "income": 5000.00,
      "expense": 0.00,
      "count": 1
    }
  ]
}
```

- 返回该月有交易记录的每一天的汇总
- 无交易记录的天数不返回 (前端用 0 填充)
- 该接口有缓存, TTL 5 分钟

## 3. 错误码汇总

| HTTP 状态码 | Message 示例 | 典型场景 |
|-------------|-------------|---------|
| 400 | `"Key: 'CreateTransactionInput.Amount' Error:...gt"` | 参数校验失败 |
| 400 | `"old password is incorrect"` | 修改密码时旧密码不匹配 |
| 400 | `"invalid format, must be csv or json"` | 导出格式参数不合法 |
| 400 | `"ids must not be empty"` | 批量操作未提供 ID |
| 401 | `"missing authorization header"` | 未携带 Token |
| 401 | `"invalid or expired token"` | Token 过期 |
| 401 | `"token revoked"` | Token 已被撤销 (登出后) |
| 404 | `"ledger not found"` | 账本 ID 不存在 |
| 404 | `"transaction not found"` | 交易 ID 不存在 |
| 404 | `"category not found"` | 分类 ID 不存在 |
| 409 | `"username or email already exists"` | 注册时用户名或邮箱重复 |
| 409 | `"cannot delete category with existing transactions"` | 分类下有交易记录 |
| 409 | `"email already in use"` | 修改邮箱地址已存在 |
| 500 | `"failed to create transaction"` | 数据库写入失败 |

---

## 4. API 路由总表

```
POST   /api/v1/auth/register                           # 注册
POST   /api/v1/auth/login                              # 登录
GET    /api/v1/auth/me                                  # 当前用户
POST   /api/v1/auth/logout                              # 登出
PUT    /api/v1/auth/password                            # 修改密码
PUT    /api/v1/auth/email                               # 修改邮箱

GET    /api/v1/ledgers                                  # 账本列表
POST   /api/v1/ledgers                                  # 创建账本
GET    /api/v1/ledgers/:ledger_id                       # 账本详情
PUT    /api/v1/ledgers/:ledger_id                       # 更新账本
DELETE /api/v1/ledgers/:ledger_id                       # 删除账本
GET    /api/v1/ledgers/:ledger_id/summary               # 账本汇总
GET    /api/v1/ledgers/:ledger_id/monthly-trend          # 月度趋势
GET    /api/v1/ledgers/:ledger_id/category-breakdown     # 分类分布
GET    /api/v1/ledgers/:ledger_id/daily-transactions     # 日历汇总
GET    /api/v1/ledgers/:ledger_id/export                 # 导出交易
GET    /api/v1/ledgers/:ledger_id/tags                   # 标签列表

GET    /api/v1/ledgers/:ledger_id/categories            # 分类列表（树形）
POST   /api/v1/categories                               # 创建分类
PUT    /api/v1/categories/:id                            # 更新分类
DELETE /api/v1/categories/:id                            # 删除分类

GET    /api/v1/ledgers/:ledger_id/transactions           # 交易列表（分页+筛选）
POST   /api/v1/transactions                              # 创建交易
PUT    /api/v1/transactions/:id                          # 更新交易
DELETE /api/v1/transactions/:id                          # 删除交易
POST   /api/v1/transactions/batch-delete                 # 批量删除
PUT    /api/v1/transactions/batch-update                 # 批量修改

GET    /api/v1/exchange-rates                            # 汇率列表
POST   /api/v1/exchange-rates                            # 创建汇率
GET    /api/v1/exchange-rates/latest                     # 最新汇率
DELETE /api/v1/exchange-rates/:id                        # 删除汇率

GET    /api/v1/health                                    # 健康检查
GET    /swagger/*any                                     # Swagger UI
GET    /metrics                                          # Prometheus 指标
```

总计: 33 个端点 (含 Swagger 和 Metrics), 不含 Swagger/Metrics 为 31 个业务端点
