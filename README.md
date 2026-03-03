# AskDB (Go, MySQL, OpenAI)

一个可落地的 **自然语言转 SQL** HTTP API 服务（只读模式）。

## 特性

- 基于 OpenAI Chat Completions 将自然语言转 MySQL SQL
- 从 `schema.sql` 加载完整表结构（可裁剪）
- SQL 安全防护：仅允许 `SELECT/CTE`、禁止多语句和写操作

## 目录结构

- `cmd/server/main.go`: 启动入口
- `internal/config`: 环境变量配置
- `internal/llm`: OpenAI 调用
- `internal/schema`: schema 文件加载
- `internal/sqlguard`: 只读安全校验
- `internal/service`: 业务编排（生成+校验+可选执行）
- `internal/httpapi`: HTTP 路由

## 快速开始

1. 准备 schema 文件：`schema.sql`
2. 配置环境变量：

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_MODEL="gpt-4.1-mini"
export SCHEMA_SQL_PATH="./schema.sql"
```

3. 运行：

```bash
go run ./cmd/server
```

## API

### 健康检查

`GET /healthz`

### 自然语言转 SQL

`POST /v1/nl2sql`

请求：

```json
{
  "question": "查询近30天订单数最多的10个用户",
  "max_rows": 100,
  "execute": false
}
```

响应：

```json
{
  "sql": "SELECT ... LIMIT 100",
  "reasoning": "..."
}
```

当前版本仅返回 SQL，不直接执行数据库查询。

## 安全建议

- 数据库账号务必设置为只读权限
- 建议在网关再做 SQL 审计/超时/限流
- 生产可加入 AST 级校验（当前为规则校验）
