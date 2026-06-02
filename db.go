package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

type AccessLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	IP        string    `gorm:"index;size:45" json:"ip"`
	UserAgent string    `gorm:"size:512" json:"user_agent"`
	Path      string    `gorm:"size:256" json:"path"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

var (
	db        *gorm.DB
	startTime = time.Now()
)

func InitDB(d *gorm.DB) {
	db = d
	db.AutoMigrate(&AccessLog{})
}

func GetDB() *gorm.DB {
	return db
}

func GetStartTime() time.Time {
	return startTime
}

func GetAccessLogs() []AccessLog {
	if db == nil {
		return nil
	}
	var logs []AccessLog
	db.Order("last_seen desc").Find(&logs)
	return logs
}

func GetIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	if fwd := r.Header.Get("X-Real-IP"); fwd != "" {
		return fwd
	}
	idx := stringsLastIndex(r.RemoteAddr, ":")
	if idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

func stringsLastIndex(s string, sep byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return i
		}
	}
	return -1
}
