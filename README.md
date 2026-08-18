# 邻里同行：社区志愿服务小队协作平台

项目编号：`GO-CG-026`；原始分类：代码生成；来源：`word2.xlsx / Sheet1 / 第 31 行`。

邻里同行是一套可离线运行的社区志愿协作系统。居民可以发现公开小队、提交加入申请、认领岗位、签到和登记服务；负责人可以复核时长、处理更正、交接现场风险并在工作台查看缺岗和逾期事项。它不提供实时聊天，也不依赖第三方推送、地图、云存储或在线模型。

## 模块职责

- `cmd/server`：进程装配、HTTP 服务和优雅停机。
- `internal/domain`：成员任期、活动岗位、认领、服务记录版本、风险与交接领域模型。
- `internal/application`：按业务闭环拆分的用例，接口定义靠近调用方，统一使用 context 和乐观事务。
- `internal/repository`：并发安全的内存仓储、PostgreSQL JSONB 聚合快照与不可变审计事务、确定性演示数据。
- `internal/transport/http`：Gin 路由、请求校验、统一错误和 API DTO。
- `internal/middleware`：request ID、超时、CORS、安全响应头、恢复和脱敏访问日志。
- `internal/service`：本地通知、受控附件、时钟和定时任务适配器。
- `migrations`：PostgreSQL 表、约束、索引和可重复种子 SQL。
- `api/openapi`：OpenAPI 3.0 接口契约。
- `web`：Vue 3、TypeScript、Vite、Pinia 中文工作台。

## 业务闭环

1. 居民从社区发现页查看公开小队，公开招募可直接申请，私有小队需邀请码。
2. 负责人审批申请；暂停、退出和队长移交都会产生成员事件，队长未移交不能退出。
3. 负责人创建活动、岗位和容量并发布。成员认领时系统同时检查活动状态、当前成员资格、岗位容量及个人时间冲突；满员后进入有序候补。
4. 认领成员完成到场核验后开始服务，结束时间由服务端时钟记录，跨度必须在 1 分钟到 24 小时之间。
5. 负责人复核服务记录。已确认记录不会被覆盖，更正会生成更高版本，审批前原版本仍有效。
6. 活动中的风险可以拆分跟进项并交接给另一名成员。未解除风险没有有效交接时，活动不能完成。
7. 站内提醒、个人历史、CSV 服务证明和负责人工作台都从同一事务状态读取。

状态摘要：

- 小队：`recruiting -> paused -> closed`
- 成员：`pending -> active <-> paused -> exited`
- 活动：`draft -> published -> completed`，发布前后均可转为 `cancelled`
- 认领：`active | waitlisted -> checked_in | cancelled | no_show`
- 服务：`started -> pending_review -> confirmed`；更正产生新的 `pending_review` 版本
- 风险：`open -> in_progress -> resolved`
- 交接：`pending -> acknowledged -> completed`，读取工作台时逾期项标识为 `overdue`

## 本地启动

要求 Go 1.24+、Node.js 22+；PostgreSQL 17 仅在持久化模式需要。

```powershell
Copy-Item .env.example .env
$env:STORE_MODE = 'memory'
go run ./cmd/server
```

另开终端启动前端：

```powershell
Set-Location web
npm install
npm run dev
```

打开 `http://127.0.0.1:5173`。内存模式预置两个社区、两个小队、四名用户、计划中与已完成活动、待处理高风险、过期交接和公告。

演示身份通过页面左侧切换，本地请求对应：

- 成员：`X-User-ID: u-volunteer`、`X-User-Role: member`
- 负责人：`X-User-ID: u-captain`、`X-User-Role: captain`
- 管理员：`X-User-ID: u-admin`、`X-User-Role: admin`

这些 Header 只用于离线演示适配器，不应直接暴露到公网。

## PostgreSQL 与迁移

```powershell
$env:DATABASE_URL = 'postgres://volunteer:volunteer@localhost:5432/volunteer?sslmode=disable'
./scripts/migrate.ps1
$env:STORE_MODE = 'postgres'
$env:SEED_DEMO = 'true'
go run ./cmd/server
```

迁移全部使用 `IF NOT EXISTS` 或 `ON CONFLICT DO NOTHING`，重复执行不会覆盖用户数据。生产仓储在一笔 `SERIALIZABLE` 事务中锁定聚合版本、写 JSONB 快照并追加审计事件；陈旧版本返回 409。

也可执行：

```powershell
docker compose up --build
```

容器入口为 `http://localhost:8088`。镜像使用官方多架构基础镜像，可在 amd64 或 arm64 主机原生构建。

## 配置

配置项见 `.env.example`。关键项包括 `STORE_MODE`、`DATABASE_URL`、`ATTACHMENT_DIR`、`MAX_UPLOAD_BYTES` 和 `REQUEST_TIMEOUT`。附件适配器只接受 JPG、PNG、PDF，拒绝目录穿越和超限文件；数据库及日志不记录附件正文、令牌或密码。

## API 示例

发现小队：

```powershell
curl.exe 'http://localhost:8080/api/v1/discovery/teams?page=1&page_size=20&status=recruiting'
```

负责人创建活动：

```powershell
curl.exe -X POST http://localhost:8080/api/v1/activities `
  -H 'Content-Type: application/json' -H 'X-User-ID: u-captain' -H 'X-User-Role: captain' `
  -d '{"team_id":"team-riverside","title":"楼道安全巡查","start_at":"2026-08-22T09:00:00Z","end_at":"2026-08-22T11:00:00Z","location":"滨河社区"}'
```

完整接口见 `api/openapi/openapi.yaml`。所有业务 API 位于 `/api/v1`；错误均包含稳定 `code`、中文 `message`、字段错误及 `request_id`。

## 测试与验证

```powershell
go build ./...
go test ./...
go test -race ./...
go vet ./...
Set-Location web
npm test
npm run build
```

2026-08-17 的最终基线验证中，`go build ./...`、`go test ./...`、`go test -race ./...` 和 `go vet ./...` 均通过；前端 `npm test` 与 `npm run build` 均通过。测试覆盖活动取消后禁止认领、满员时拒绝候补提升、容量并发、队长移交、服务记录更正、内存仓储乐观事务、附件路径限制以及 HTTP 权限和校验错误结构。

## 健康与审计

- `GET /healthz`：进程存活。
- `GET /readyz`：服务已完成仓储装配并可接收请求。
- 每个请求生成或透传 `X-Request-ID`。
- 写操作将 actor、request ID 和时间写入不可变审计流；PostgreSQL 模式与业务快照同事务提交。
- 服务收到 Ctrl+C 或 SIGTERM 后停止就绪并在 10 秒内优雅关闭。
