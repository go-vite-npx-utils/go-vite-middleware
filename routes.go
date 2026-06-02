package middleware

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router) {
	r.Get("/api/debug/info", DebugInfo)
	r.Get("/api/debug/logs", AccessLogsPage)
	r.Post("/api/debug/logs/clear", ClearAccessLog)
	r.Get("/api/debug/turnout", TurnoutHandler)

	r.Route("/api/sys", func(r chi.Router) {
		r.Get("/pulse", SysPulse)
		r.Get("/fetch", SysFetch)
		r.Get("/lookup", SysLookup)
		r.Get("/query", SysQuery)
		r.Get("/console", SysConsole)
	})

	r.Get("/console", ServeConsole)
	r.Get("/kiosk/carnival", ServeCarnival)
}
