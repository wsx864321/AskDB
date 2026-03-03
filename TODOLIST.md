# AskDB TODO List

> 目标：为 NL2SQL 服务建立可持续演进路线，按“效果优先、风险可控、逐步上线”推进。

## P0（优先立即做）

- [ ] 建立离线评测集（至少 100 条）
  - [ ] `question -> gold tables -> gold sql` 数据格式
  - [ ] 指标：Recall@K、SQL 可执行率、语义正确率
  - [ ] 每次改动后自动回归

- [ ] 检索二阶段优化（在现有 Hybrid 召回后增加 rerank）
  - [ ] 第一阶段：Hybrid（Lexical + BM25）召回 TopN
  - [ ] 第二阶段：字段命中、时间字段偏好、join key 命中重排
  - [ ] 最终截断 TopK 并对比离线评测收益

- [ ] SQL Guard 升级为 AST 解析校验
  - [ ] 仅允许 SELECT/CTE
  - [ ] 子查询同样执行只读检查
  - [ ] 表白名单 / schema 白名单（可配置）
  - [ ] LIMIT、复杂度、函数黑名单更严格控制

## P1（第二阶段）

- [ ] 增加 Embedding 融合召回（先不引入向量库）
  - [ ] 离线/内存向量索引
  - [ ] 融合打分：Lexical + BM25 + Embedding
  - [ ] 对比评测收益与 token 成本

- [ ] Prompt 精细化
  - [ ] 按问题类型路由模板（聚合/排行/明细）
  - [ ] few-shot 动态选择（非全量注入）
  - [ ] 修复 Prompt 增强：附加失败类别与 SQL 片段

- [ ] 观测与成本治理
  - [ ] 记录 `recalled_tables`、`attempts`、召回耗时、LLM耗时
  - [ ] 记录 prompt/completion token 与成本
  - [ ] 统计 guard 失败原因分布

## P2（上线治理）

- [ ] 安全与权限
  - [ ] 网关鉴权、限流、超时、审计日志
  - [ ] 敏感列脱敏/拦截策略
  - [ ] 多租户隔离规则

- [ ] 发布与回滚
  - [ ] prompt/retriever/guard 版本化
  - [ ] 灰度开关与 A/B 策略
  - [ ] 快速回滚机制

- [ ] 执行闭环（如未来需要执行 SQL）
  - [ ] EXPLAIN 成本阈值拦截
  - [ ] 查询超时 + 返回行数硬限制
  - [ ] 只读数据库账户最小权限

## 文档与协作

- [ ] 增加 `docs/evaluation.md`：评测样本格式与打分方法
- [ ] 增加 `docs/retrieval.md`：召回架构与参数说明
- [ ] 增加 `docs/guard.md`：安全策略与风险案例
- [ ] PR 模板补充：必须附回归指标对比
