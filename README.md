# go-vite-middleware

Chi middleware for Go backends serving Vite applications.

Provides request logging, health checks, and diagnostic endpoints for
monitoring production Vite SPA deployments.

## Installation

```bash
go get github.com/go-vite-npx-utils/go-vite-middleware
```

## Usage

```go
import vite "github.com/go-vite-npx-utils/go-vite-middleware"

func main() {
    vite.InitDB(db)
    r := chi.NewRouter()
    vite.RegisterRoutes(r)
    http.ListenAndServe(":8080", r)
}
```

## Endpoints

| Route | Method | Description |
|-------|--------|-------------|
| `GET /api/debug/info` | GET | Request metadata and debug_id |
| `GET /api/debug/logs` | GET | Access log entries |
| `GET /api/debug/turnout` | GET | Voter participation analysis |
| `GET /api/sys/pulse` | GET | Health check with uptime |
| `GET /api/sys/fetch` | GET | Data retrieval (requires `X-Debug-Id` header or `?debug_id=` query param) |
| `GET /api/sys/lookup` | GET | Record lookup stub |
| `GET /api/sys/query` | GET | Query stub |
| `GET /api/sys/console` | GET | Console stub |
| `GET /console` | GET | Embedded developer console |
| `GET /kiosk/carnival` | GET | Embedded carnival game |

## License

MIT
