# AskDB (Go, MySQL, OpenAI)

一个面向 **MySQL 只读场景** 的自然语言转 SQL（NL2SQL）HTTP API 服务。

> 核心思路：**表召回（Schema Retrieval）→ LLM 生成 SQL → 只读安全校验 →（可选）LLM 自动修复**。

## 适用场景

- 表数量较多（例如 100+）的业务库，不适合把全量 schema 直接喂给模型。
- 只需要 SQL 生成，不希望服务直接执行数据库写操作。
- 需要可观测、可调优的提示词上下文（schema + glossary + few-shot）。

## 核心能力

- **MySQL NL2SQL**：基于 OpenAI Chat Completions 生成 SQL。
- **表召回（重点）**：从完整 schema 中先召回相关表，再把召回结果提供给模型。
- **术语映射**：支持 `glossary.md`（业务术语 -> 表/字段）增强召回与生成。
- **Few-shot 示例**：支持 `fewshot.jsonl` 提升 SQL 风格与准确率。
- **只读安全**：仅允许 `SELECT/CTE`、禁止多语句和危险关键词、自动 LIMIT 控制。
- **自动修复**：首轮 SQL 不通过 guard 时，基于违规原因触发修复重试。

## 工作流

1. 读取完整 `schema.sql` 并解析 `CREATE TABLE ...` 语句。
2. 对用户问题（结合 glossary）做词项匹配与打分，召回 Top-K 相关表。
3. 将“召回表结构 + glossary + few-shot + 问题”送入 LLM 生成 SQL。
4. 使用 SQL guard 做只读校验与 LIMIT 收敛。
5. 若校验失败，在配置次数内触发 LLM 修复并再次校验。
6. 返回最终 SQL（以及 attempts / recalled_tables）。

## 项目结构

- `cmd/server/main.go`：服务入口
- `internal/httpapi`：HTTP 路由与请求处理
- `internal/service`：NL2SQL 主流程编排
- `internal/retrieval`：表召回逻辑
- `internal/llm`：OpenAI 调用与 prompt 组装
- `internal/sqlguard`：只读 SQL 安全校验
- `internal/schema`：schema / glossary / few-shot 加载与表提取
- `schema.sql`：数据库表结构输入
- `glossary.md`：可选术语映射
- `fewshot.jsonl`：可选示例集

## 快速开始

### 1) 准备文件

- 必需：`schema.sql`
- 可选：
  - `glossary.md`（建议有）
  - `fewshot.jsonl`（建议有）

`fewshot.jsonl` 格式（每行一个 JSON）：

```json
{"question":"统计近7天下单用户数","sql":"SELECT COUNT(DISTINCT user_id) FROM orders WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)"}
```

### 2) 配置环境变量

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4.1-mini"

export SCHEMA_SQL_PATH="./schema.sql"
export GLOSSARY_PATH="./glossary.md"
export FEWSHOT_PATH="./fewshot.jsonl"

# 返回行数控制
export DEFAULT_ROW_LIMIT="100"
export MAX_ROW_LIMIT="1000"

# guard 失败后的修复重试次数
export GUARD_REPAIR_TRIES="2"

# 表召回参数（100+表场景建议重点调优）
export RECALL_TOP_K="12"
export RECALL_MAX_BYTES="60000"
export RECALL_LEXICAL_WEIGHT="1.0"
export RECALL_BM25_WEIGHT="1.0"
export RECALL_NAME_BOOST="8.0"
```

### 3) 启动服务

```bash
go run ./cmd/server
```

默认监听：`http://localhost:8080`

## API

### `GET /healthz`

健康检查。

### `POST /v1/nl2sql`

请求体：

```json
{
  "question": "查询近30天订单数最多的10个用户",
  "max_rows": 100,
  "execute": false
}
```

> 说明：当前版本 `execute` 仅保留字段兼容，服务不会直接执行 SQL。

响应体示例：

```json
{
  "sql": "SELECT user_id, COUNT(*) AS cnt FROM orders WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) GROUP BY user_id ORDER BY cnt DESC LIMIT 10",
  "reasoning": "根据订单表按近30天聚合用户下单次数并排序",
  "attempts": 1,
  "recalled_tables": 6
}
```

字段说明：

- `sql`：最终通过只读校验后的 SQL。
- `reasoning`：模型简要推理说明。
- `attempts`：生成+修复总尝试次数。
- `recalled_tables`：本次进入 prompt 的召回表数量。

## 调优建议（重点）

- **召回不准**：
  - 增加 `glossary.md` 中业务别名与字段映射。
  - 提高 `RECALL_TOP_K`（如 12 -> 20）。
- **SQL 漏字段/错表**：
  - 在 `fewshot.jsonl` 增加高频复杂问法样例。
  - 保证 schema 中字段注释语义清晰。
- **token 成本高**：
  - 降低 `RECALL_TOP_K` 或 `RECALL_MAX_BYTES`。
  - 精简 few-shot 到最有效样例。

## 安全建议

- 生产环境数据库账号应为 **只读权限**。
- 在网关侧增加：鉴权、限流、超时、审计日志。
- 如需更强约束，可在后续增加 AST 级 SQL 解析与白名单策略。

## 当前限制

- 仅支持 MySQL 方言目标。
- 表召回当前为 Hybrid（关键词 + BM25）评分（轻量、无外部依赖），后续可升级 Embedding 融合召回。
- 当前仅返回 SQL，不执行数据库查询。

## 协作说明（分支冲突处理）

- 若历史分支出现冲突，建议从最新主干重新拉分支后再提交，避免重复冲突。
- 提交前可先 `git rebase`（或 `merge`）同步主干，再执行测试与提交。
- 本仓库建议保持 PR 小步快跑：功能变更与文档变更尽量拆分。
