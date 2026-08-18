# 独立语义覆盖复核（cry031 baseline）

复核依据是 `_shared/requirements/cry031/source-prompt.txt` 和锁定 manifest；本文只记录 baseline 的可观察实现证据，不替代 `requirements.lock.json`，也不包含任何埋题答案。

| Requirement | 生产实现证据 | 行为测试证据 | 数据库/前端证据 | 覆盖结论 |
|---|---|---|---|---|
| REQ-001 项目身份 | `README.md` 项目标题与模块入口 | `tests/http_test.go` 健康接口 | `web/src/App.vue` | pass |
| REQ-002 GO-CG-026 编号 | `README.md` 交付说明 | 无需运行时断言 | `api/openapi/openapi.yaml` info | pass |
| REQ-003 代码生成分类 | `README.md` baseline 说明 | 无需运行时断言 | `go.mod`、`web/package.json` | pass |
| REQ-004 来源记录 | `README.md` 需求说明 | 无需运行时断言 | contract 外部锁定 | pass |
| REQ-005 相互关联模块 | `internal/application/{membership,activities,service_records,collaboration}.go` | `internal/application/application_test.go` 跨闭环用例 | `migrations/001_init.sql` 外键与索引 | pass |
| REQ-006 社区/小队/领域/公开范围/招募 | `internal/domain/models.go`、`application/catalog.go`、`membership.go` | `tests/http_test.go` discovery | `migrations/001_init.sql` communities/service_areas/teams；`web/src/pages/DiscoveryPage.vue`、`TeamPage.vue` | pass |
| REQ-007 申请/邀请码/队长移交/暂停/退出/轨迹 | `application/membership.go`、`domain/models.go` | `TestCaptainTransferKeepsSingleLeader`、`TestMembershipPauseCanBeResumed` | `membership_events` 表；Team 页面 | pass |
| REQ-008 活动排期/岗位/时段/容量/地点/物资 | `application/activities.go` | `TestConcurrentClaimsRespectCapacity` | `activities`、`role_slots` 表；`ActivityPage.vue` | pass |
| REQ-009 认领/候补/取消/签到/冲突防护 | `application/activities.go` | `TestCancelledActivityCannotBeClaimedAndCancelsWaitlist`、`TestPromoteWaitlistDoesNotExceedCapacity`、并发容量测试 | `claims` 索引；活动详情页 | pass |
| REQ-010 开始/结束/更正/照片说明/复核 | `application/service_records.go` | `TestCorrectionPreservesHistoryUntilApproved`、时长领域测试 | `service_records` 版本约束；`ServiceRecordsPage.vue` | pass |
| REQ-011 风险/交接/跟进 | `application/collaboration.go` | `TestRiskHandoffBlocksResolutionUntilAcknowledged` | `risks`、`handoffs`、`follow_ups` 表；`HandoffPage.vue` | pass |
| REQ-012 公告/站内提醒且无外部聊天 | `application/collaboration.go`、`service/notifier.go` | `TestAnnouncementCreatesLocalNotice` | notices/announcements 表；Team 页面 | pass |
| REQ-013 个人档案/历史/服务证明 | `application/service_records.go` (`ExportCertificate`) | correction/export 断言 | `service_records` 版本；`ProfilePage.vue` | pass |
| REQ-014 负责人待确认/缺岗/逾期交接/复盘 | `application/collaboration.go` (`LeaderDashboard`) | 风险交接测试与 HTTP dashboard 冒烟 | dashboard 查询所用业务表；`LeaderDashboardPage.vue` | pass |
| REQ-015 取消后禁认领/确认记录只能更正 | `domain.Activity.CanClaim`、`application/activities.go`、`service_records.go` | 取消和更正测试 | 状态 CHECK 与版本唯一约束 | pass |
| REQ-016 时长限制、中文页面和异常种子 | `domain.ServiceRecord.Duration`、`repository/seed.go` | `TestServiceDurationRejectsUnreasonableSpan` | 7 个 `web/src/pages/*.vue`；`seed.go` 异常交接 | pass |
| REQ-017 Vue | `web/package.json`、`web/src/main.ts` | `npm test` 类型检查 | `App.vue` | pass |
| REQ-018 TypeScript/Vite/Pinia | `web/package.json`、`vite.config.ts`、`stores/workspace.ts` | `npm run build` | `web/src` | pass |
| REQ-019 Go 1.24+ | `go.mod`、`cmd/server/main.go` | `go build ./...` | N/A | pass |
| REQ-020 Gin/pgx/validator/zap/OpenAPI | `transport/http`、`repository/postgres.go`、`middleware`、`api/openapi` | HTTP 校验测试、`go vet` | `go.mod` | pass |
| REQ-021 PostgreSQL/离线 | `repository/postgres.go`、`service/files.go`、`service/notifier.go` | memory tests；PG test gated by `DATABASE_URL` | `migrations/`、`docker-compose.yml` | pass（需容器环境复验 PG） |
| REQ-022 本地适配器/API/分页排序/错误 request_id | `middleware`、`transport/http`、`service` | `TestValidationErrorIncludesRequestIDAndFields`、附件上传、排序拒绝 | OpenAPI paths and local file adapter | pass |
| REQ-023 README/迁移/种子/测试/无敏感产物 | `README.md`、`Makefile`、`.env.example`、`.gitignore` | `go test`/`go vet`/web build | migration/compose/scripts | pass |

## 缺口与后续复核

- PostgreSQL 真实集成测试需要在 `docker compose` 启动后，以 `DATABASE_URL` 运行；离线默认模式已用并发安全内存仓储覆盖同一乐观版本协议。
- 认证目前是本地演示 Header 适配器；生产部署必须在网关或后续身份服务中替换，应用层不会信任 Header 中的角色，仍以快照中的用户和成员角色授权。
- 真实附件哈希和数据库附件元数据表已预留，当前 baseline 返回受控相对 token；不得把绝对路径或运行期文件提交到 GitHub。
- 未实现实时聊天、外部消息、地图、云转码和在线模型，这些均是原始需求明确排除项。
