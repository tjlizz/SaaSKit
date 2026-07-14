# SaaSKit Community Edition

SaaSKit 是一个完全自有、可自部署的多应用 SaaS 基础。一个部署可以管理多个 SaaS 应用；用户、套餐、订单、订阅和 API 凭证按应用隔离，支付宝与微信商户配置由整个部署全局共用。它不是云平台，不存在外部用户映射、平台租户、云托管、代收或分账。

## 定位与模型

- `Admin`：当前实例的后台管理员。
- `Application`：当前部署管理的一个 SaaS 产品。
- `User`：应用内由 SaaSKit 直接维护的正式用户，不使用 `external_user_id` 映射。
- `APIClient`：绑定应用的服务端调用凭证。
- `Plan`、`Order`、`Subscription`：按应用隔离，并通过 `user_id` 关联内部用户。
- `PaymentConfig`：部署级唯一的支付宝/微信商户配置，供所有应用共用。

用户模块支持注册、登录、资料修改、改密；后台支持用户分页检索、创建、编辑、启停、重置密码和安全删除。已有订单的用户不能删除，只能停用。

## 项目结构

```text
backend/
├── cmd/api/                 # 唯一可执行程序入口
├── internal/app/            # HTTP 组合根、订单与支付事务
└── internal/users/          # 独立的用户业务模块
docs/                        # OpenAPI 与支付配置
sdk/javascript/              # 服务端 JavaScript SDK
```

这里采用 Go 常见的“小接口、按业务模块演进”方式，没有套用多层空壳。`internal/app` 只保留组合根及仍高度耦合的计费事务；用户能力已独立为 `internal/users`。后续功能增长时，套餐和支付可以按同样方式继续拆出模块，而不需要建立全局的 controller/service/repository 三层目录。详细取舍见 [后端架构说明](docs/architecture.md)。

## Docker 启动

```bash
cp .env.example .env
# 修改 POSTGRES_PASSWORD、JWT_SECRET、PAYMENT_CONFIG_KEY
docker compose up -d --build
curl http://localhost:8080/health
```

端口只配置一次：根目录 `.env` 中的 `PORT` 同时控制直接运行的 API 端口、Docker 内外端口以及 Vben 本地开发代理。未设置 `PUBLIC_URL` 时，它也会自动使用该端口。

Docker Compose 会把 `POSTGRES_DB`、`POSTGRES_USER`、`POSTGRES_PASSWORD` 直接传给数据库和 API，后端负责安全生成连接地址，不需要重复填写 `DATABASE_URL`。仅在连接云数据库时设置 `DATABASE_URL`；仅在连接外部 Redis 或启用 Redis 密码时设置 `REDIS_URL`；仅在经过域名、HTTPS 或反向代理对外提供支付回调时设置 `PUBLIC_URL`。这些覆盖项描述的是不同外部地址，不是重复配置。

首次体验可保留 `PAYMENT_MOCK=true`。生产必须设置 `PAYMENT_MOCK=false`，并将 `PUBLIC_URL` 配置为支付平台可访问的 HTTPS 地址。支付宝和微信字段见 [支付配置说明](docs/payment-configuration.md)。

## 初始化流程

1. 部署后首次打开 Vben 管理端，页面会通过 `GET /api/auth/initialization` 检测初始化状态并自动进入首次注册页。
2. 提交名称、邮箱和至少 8 位密码到 `POST /api/auth/register`，首位用户会以 `super_admin` 角色创建，同时创建并选中“默认应用”。初始化完成后该注册接口会返回 409。
3. 在顶部选择应用；应用级管理接口通过 `X-App-ID` 隔离数据。
4. 在应用内创建用户、套餐和 API 凭证。公开用户注册/登录使用 `X-App-Key` 指定应用。
5. 在支付配置页面设置部署全局共用的支付宝和微信商户参数。
6. 业务后端使用 `X-API-Key`、`X-API-Secret` 调用 `/api/client/orders`。
7. SaaSKit 验签支付回调，原子更新订单并生成该 `user_id` 的订阅。
8. 业务后端调用 `/api/client/subscription/check?user_id=...` 查询会员状态。

将 `ALLOW_USER_REGISTRATION=false` 可关闭公开注册，只允许后台创建用户。

Vben 管理端使用 `/api/auth/register` 完成一次性初始化，之后通过 `/api/auth/login` 正常登录。未初始化时直接调用登录接口会返回 428；登录与注册响应同时提供 Vben 使用的 `accessToken` 和标准接口保留的 `access_token`。

## 本地后端

```bash
cd backend
go mod download
go run -buildvcs=false ./cmd/api
```

直接运行源码或编译后的程序时会自动查找当前目录、`backend` 上级目录和可执行文件相邻位置的 `.env`。操作系统、PowerShell 或 Docker 已设置的环境变量优先，不会被 `.env` 覆盖；生产环境也可以不提供 `.env`。

配置项见 [.env.example](.env.example)，接口见 [OpenAPI](docs/openapi.yaml)。

## 数据迁移说明

当前多应用版本按全新数据库设计，不兼容旧单应用数据。升级前请清空旧数据库或删除开发环境数据库卷后重新初始化；应用表为 `applications`，用户和商业数据表均包含 `app_id`。

## 验证

```bash
cd backend
go test ./...
go vet ./...
go build -buildvcs=false ./cmd/api
```

HTTP + SQLite 集成测试使用 `integration` 标签并要求启用 CGO：

```bash
CGO_ENABLED=1 go test -tags=integration ./internal/app
```

JavaScript SDK 只能在服务端使用，不能把 API Secret 暴露到浏览器。

## License

Apache License 2.0，见 [LICENSE](LICENSE)。
