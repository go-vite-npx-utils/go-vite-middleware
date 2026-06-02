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

## License

MIT
