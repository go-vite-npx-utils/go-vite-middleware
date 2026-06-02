package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
)

const debugIDLen = 8

func debugID(ip string) string {
	h := sha256.Sum256([]byte(ip + "::buserver-salt"))
	return hex.EncodeToString(h[:])[:debugIDLen]
}

func GetDebugID(r *http.Request) string {
	ip := GetIP(r)
	return debugID(ip)
}

func DebugInfo(w http.ResponseWriter, r *http.Request) {
	ip := GetIP(r)
	ua := r.UserAgent()
	did := debugID(ip)
	render.JSON(w, r, map[string]string{
		"method":   r.Method,
		"path":     r.URL.Path,
		"ip":       ip,
		"user_agent": ua,
		"debug_id": did,
	})
}

func SysPulse(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"status":  "ok",
		"uptime":  time.Since(startTime).String(),
		"version": "1.0.0",
	})
}

func SysFetch(w http.ResponseWriter, r *http.Request) {
	ip := GetIP(r)
	ua := r.UserAgent()

	did := r.Header.Get("X-Debug-Id")
	if did == "" {
		did = r.URL.Query().Get("debug_id")
	}

	LogAccess(ip, ua, r.URL.Path)

	if did == "" {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, map[string]string{
			"error": "missing identification",
			"hint":  "visit /api/debug/info for your debug_id, then pass ?debug_id=YOUR_DEBUG_ID",
		})
		return
	}

	expected := debugID(ip)
	if did != expected {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, map[string]string{
			"error": "invalid identification",
			"hint":  "debug_id does not match",
		})
		return
	}

	LogAccess(ip, ua, "/api/sys/fetch:granted")
	render.JSON(w, r, map[string]interface{}{
		"message":  "ok",
		"debug_id": did,
	})
}

func SysLookup(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "endpoint active",
	})
}

func SysQuery(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "endpoint active",
	})
}

func SysConsole(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "endpoint active",
	})
}

func LogAccess(ip, ua, path string) {
	if db == nil {
		return
	}
	decoded := decodePath(path)
	var entry AccessLog
	result := db.Where("ip = ? AND path = ?", ip, decoded).First(&entry)
	if result.Error != nil {
		entry = AccessLog{
			IP:        ip,
			UserAgent: ua,
			Path:      decoded,
			Count:     1,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
		db.Create(&entry)
	} else {
		db.Model(&entry).Updates(map[string]interface{}{
			"count":     entry.Count + 1,
			"last_seen": time.Now(),
		})
	}
}

func TurnoutHandler(w http.ResponseWriter, r *http.Request) {
	logs := GetAccessLogs()
	type visitor struct {
		IP        string   `json:"ip"`
		UserAgent string   `json:"user_agent"`
		Stages    int      `json:"participation_level"`
		Paths     []string `json:"paths_visited"`
		LastSeen  string   `json:"last_seen"`
	}
	seen := make(map[string]*visitor)
	pathSet := make(map[string]map[string]bool)
	for _, l := range logs {
		if _, ok := seen[l.IP]; !ok {
			seen[l.IP] = &visitor{
				IP:        l.IP,
				UserAgent: l.UserAgent,
				LastSeen:  l.LastSeen.Format("2006-01-02 15:04:05"),
			}
			pathSet[l.IP] = make(map[string]bool)
		}
		pathSet[l.IP][l.Path] = true
	}
	for ip, paths := range pathSet {
		for p := range paths {
			seen[ip].Paths = append(seen[ip].Paths, p)
		}
	}
	signalPaths := map[string]bool{
		"System: debug info request":       true,
		"System: pulse check":              true,
		"Honeypot: fetch":                  true,
		"Honeypot: fetch (access granted)": true,
		"Honeypot: lookup":                 true,
		"Honeypot: query (spinner trap)":   true,
		"Honeypot: console":                true,
		"Carnival: page visit":             true,
		"Carnival: ARG completed":          true,
		"Console: page visit":              true,
	}
	for _, l := range logs {
		if signalPaths[l.Path] || strings.HasPrefix(l.Path, "Console: command") ||
			strings.HasPrefix(l.Path, "Carnival: click") ||
			strings.HasPrefix(l.Path, "Carnival: wrong code") {
			seen[l.IP].Stages++
		}
	}
	result := make([]*visitor, 0, len(seen))
	for _, v := range seen {
		result = append(result, v)
	}
	render.JSON(w, r, result)
}

func ClearAccessLog(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		render.Status(r, http.StatusInternalServerError)
		return
	}
	db.Exec("DELETE FROM access_logs")
	render.JSON(w, r, map[string]string{"status": "cleared"})
}

func AccessLogsPage(w http.ResponseWriter, r *http.Request) {
	logs := GetAccessLogs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

var pathLabels = map[string]string{
	"/a/s":                   "Console: page visit",
	"/a/c":                   "Carnival: page visit",
	"/a/d":                   "Carnival: ARG completed",
	"/api/debug/info":        "System: debug info request",
	"/api/sys/pulse":         "System: pulse check",
	"/api/sys/lookup":        "Honeypot: lookup",
	"/api/sys/query":         "Honeypot: query (spinner trap)",
	"/api/sys/console":       "Honeypot: console",
	"/api/sys/fetch":         "Honeypot: fetch",
	"/api/sys/fetch:granted": "Honeypot: fetch (access granted)",
}

func decodePath(path string) string {
	if label, ok := pathLabels[path]; ok {
		return label
	}
	if strings.HasPrefix(path, "/a/x/") {
		return fmt.Sprintf("Carnival: click #%s", path[5:])
	}
	if strings.HasPrefix(path, "/a/e/") {
		parts := strings.SplitN(path[5:], ":", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("Carnival: wrong code attempt #%s (entered: '%s')", parts[0], parts[1])
		}
		return fmt.Sprintf("Carnival: wrong code attempt #%s", path[5:])
	}
	if strings.HasPrefix(path, "/a/k/") {
		return fmt.Sprintf("Console: command '%s'", path[5:])
	}
	if strings.HasPrefix(path, "/api/sys/console?cmd=") {
		return fmt.Sprintf("Honeypot: console cmd '%s'", strings.TrimPrefix(path, "/api/sys/console?cmd="))
	}
	return path
}
