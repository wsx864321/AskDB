# AskDB (Go, MySQL, OpenAI)

一个可落地的 **自然语言转 SQL** HTTP API 服务（只读模式）。

## 第二阶段优化（本次）

- 术语映射：可选加载 `glossary.md`（业务名词 -> 表/字段映射）
- Few-shot：可选加载 `fewshot.jsonl`（问题/SQL样例）
- Guard 失败自动修复：模型首轮 SQL 触发只读规则时，自动根据违规原因重试修复
- 更严格只读防护：注释剥离、禁用 `into/outfile/load_file` 等高风险关键字
- 结构化输出约束：使用 `response_format=json_object`

## 特性

- 基于 OpenAI Chat Completions 将自然语言转 MySQL SQL
- 从 `schema.sql` 加载完整表结构（可裁剪）
- SQL 安全防护：仅允许 `SELECT/CTE`、禁止多语句和写操作

## 目录结构

- `cmd/server/main.go`: 启动入口
- `internal/config`: 环境变量配置
- `internal/llm`: OpenAI 调用与 prompt 组装
- `internal/schema`: schema / glossary / few-shot 加载
- `internal/sqlguard`: 只读安全校验
- `internal/service`: 业务编排（生成+校验+修复）
- `internal/httpapi`: HTTP 路由

## 快速开始

1. 准备 schema 文件：`schema.sql`
2. 可选准备：
   - `glossary.md`（术语说明）
   - `fewshot.jsonl`（每行一个 JSON：`{"question":"...","sql":"..."}`）
3. 配置环境变量：

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_MODEL="gpt-4.1-mini"
export SCHEMA_SQL_PATH="./schema.sql"
export GLOSSARY_PATH="./glossary.md"
export FEWSHOT_PATH="./fewshot.jsonl"
export GUARD_REPAIR_TRIES="2"
```

4. 运行：

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
  "reasoning": "...",
  "attempts": 1
}
```

当前版本仅返回 SQL，不直接执行数据库查询。

## 安全建议

- 数据库账号务必设置为只读权限
- 建议在网关再做 SQL 审计/超时/限流
- 生产可加入 AST 级校验（当前为规则校验）
