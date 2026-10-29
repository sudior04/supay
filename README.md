# Supay — Local-first Commerce & Wallet

Monorepo theo [roadmap](docs/Fintech_Ecommerce_Local_First_Roadmap.md): Go/Gin, PostgreSQL, Next.js, NATS JetStream, Docker Compose, NGINX.

## Trạng thái hiện tại

Đã có đăng ký/đăng nhập bằng mật khẩu, đổi/reset password/PIN, refresh/logout, ví VND/thẻ ảo demo và catalog (category, product, variant, giá VND, tồn kho có version/audit). PIN 6 số xác nhận từng thao tác ghi product/order, không yêu cầu khi login hoặc đọc dữ liệu. Portal ADMIN quản lý category/sản phẩm; MERCHANT quản lý sản phẩm của mình. Order CRUD, ledger/payment, merchant onboarding, event publisher/consumer và observability chuyên sâu vẫn nằm trong [implementation plan](docs/Implementation_Plan.md).

API và cURL: [PIN](docs/api/pin.md), [CRUD category/sản phẩm/variant/tồn kho](docs/api/catalog.md), [lệnh Docker](docs/docker/README.md).

## Chạy toàn stack

Cài Docker Desktop với Linux containers + Compose v2. Không cần cài Go/Node để chạy Docker.

```powershell
if (!(Test-Path .env)) { Copy-Item .env.example .env }
docker compose config --quiet
docker compose up -d --build
docker compose ps -a
```

Mở http://localhost:8080. Migration chạy tự động sau khi database healthy; service chỉ start khi migration hoàn tất. Container migrate thoát code 0 là bình thường. Giữ nguyên file roadmap ban đầu.

Stack local seed sẵn các tài khoản demo: `customer@gmail.com` / `customer@123`, `merchant@gmail.com` / `merchant@123`, `admin@gmail.com` / `admin@123`. Đây là credential yếu chỉ dành cho local demo; không dùng các mật khẩu này trong môi trường thật.

Không có PIN demo mặc định: user đăng nhập bình thường rồi tự thiết lập PIN 6 số trong bảo mật tài khoản trước khi ghi product/order. Migration User v6 bỏ login challenge, giữ nguyên PIN đã cấu hình. `PIN_PEPPER` riêng của User service được cấu hình qua `.env`/Compose.

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/api/v1/users/status
Invoke-RestMethod http://localhost:8080/api/v1/products/status
Invoke-RestMethod http://localhost:8080/api/v1/wallets/status
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

| Thành phần | Local port |
|---|---|
| NGINX / portal + API cùng origin | 8080 |
| Portal trực tiếp | 3000 |
| Gateway trực tiếp | 8084 |
| User / Commerce / Wallet | 8081 / 8082 / 8083 |
| PostgreSQL User / Commerce / Wallet | 5433 / 5434 / 5435 |
| NATS / monitoring | 4222 / 8222 |

Các port chỉ bind 127.0.0.1. URL DB trong container dùng service name. Credential local có trong .env.example; đổi password sau khi DB đã có volume cần đổi cả credential trong DB hoặc reset dữ liệu demo. Auth hỗ trợ đăng ký customer, đăng nhập, refresh/logout session; các API nghiệp vụ qua gateway yêu cầu access token JWT. Đặt `JWT_SECRET` riêng, tối thiểu 32 byte, trước khi triển khai ngoài máy local.

## Cấu trúc

- `services/{gateway,user-service,commerce-service,wallet-service}`: executable độc lập; từng domain có internal/domain, application, repository, transport và migrations.
- `internal/platform`: config, HTTP lifecycle, logging, error, health, pgx pool; không chứa model nghiệp vụ.
- `web/portal`: Next.js App Router + TypeScript; auth/PIN, password, ví, category admin và sản phẩm merchant/admin đã nối API; order/checkout còn trong roadmap.
- `deployments`: Dockerfile dùng chung cho Go và cấu hình NGINX.
- `docs`: kiến trúc, ERD nháp, API convention, sequence, ADR và plan.

Dùng một Go module để giảm quản lý version local, vẫn build từng service riêng. Chỉ dùng chung code hạ tầng; không import internal của service khác.

## Phát triển ngoài Docker

Yêu cầu Go 1.26+ và Node.js 24 LTS. Dependencies được khóa bằng go.sum/package-lock.json.

```powershell
go mod download
go test ./...
go vet ./...
npm --prefix web/portal ci
npm --prefix web/portal run typecheck
npm --prefix web/portal run build
docker compose up -d user-db commerce-db wallet-db nats
powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 migrate
```

Ví dụ chạy User ở terminal riêng (khớp .env nếu đã đổi password):

```powershell
$env:HTTP_ADDR = ":8081"
$env:DATABASE_URL = "postgres://user:supay_user_local@localhost:5433/user?sslmode=disable"
$env:JWT_SECRET = "supay-local-development-secret-change-before-deploy-2026"
$env:PIN_PEPPER = "supay-local-pin-pepper-change-before-deploy-2026"
$env:COOKIE_SECURE = "false"
go run ./services/user-service/cmd/api
```

Commerce và Wallet tương tự ở 8082/8083, DB 5434/5435. Gateway dùng default upstream localhost:8081/8082/8083; chạy ở 8084:

```powershell
$env:HTTP_ADDR = ":8084"
go run ./services/gateway/cmd/api
```

Để phát triển portal cùng API origin, giữ NGINX trong Docker hoặc tự cấu hình proxy /api về gateway. Chạy portal trực tiếp bằng `npm --prefix web/portal run dev` chỉ phục vụ UI; link API cùng origin cần qua NGINX.

## Lệnh vận hành

`powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 up|down|logs|check|migrate|test|reset`; có Makefile tương ứng. Reset yêu cầu gõ RESET và xóa toàn bộ dữ liệu demo của stack. Migration mới dùng cặp `000002_name.up.sql` / `.down.sql`; không sửa migration đã áp dụng. User schema có auth/password/PIN/session/audit; Wallet có ví VND và thẻ ảo demo; Commerce có category/product/variant/inventory/audit. User v6, Commerce v2, Wallet v2.

## Giới hạn

Chưa có tiền thật, endpoint thanh toán hay tích hợp ngân hàng. Auth hiện chỉ cho phép đăng ký CUSTOMER; quy trình onboarding/duyệt MERCHANT và phân quyền theo từng nghiệp vụ chưa có. Gateway có JWT authentication nhưng chưa có rate limit. Chưa có outbox consumer, reconciliation, OTel, Prometheus/Grafana. Không coi status endpoint là bằng chứng nghiệp vụ hoặc toàn bộ dependency đã sẵn sàng: /readyz của domain kiểm tra DB; /readyz gateway chỉ kiểm tra bản thân gateway. Dùng smoke script để kiểm tra xuyên proxy.

## Phiên bản và tài liệu tham khảo

[Go releases](https://go.dev/dl/), [Gin 1.11](https://gin-gonic.com/en/blog/news/gin-1-11-0-release-announcement/), [Next.js releases](https://github.com/vercel/next.js/releases). Lockfile trong repo là nguồn phiên bản dependency chính xác.

## Kết quả kiểm tra base
[Foundation validation](docs/Foundation_Validation.md): các check đã chạy và phần Docker runtime còn cần xác minh.
