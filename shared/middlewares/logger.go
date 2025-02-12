package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type LogEntry struct {
    Method      string
    Path        string
    RemoteAddr  string
    UserAgent   string
    Status      int
    Duration    time.Duration
    TimeStamp   time.Time
}

func Logger(next http.Handler) http.Handler {
    // Logger dosyasını oluştur
    logFile, err := os.OpenFile("http.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    logger := log.New(logFile, "", log.LstdFlags)

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     
        start := time.Now()
        
   
        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
        
    
        next.ServeHTTP(ww, r)
        
     
        duration := time.Since(start)


        entry := LogEntry{
            Method:      r.Method,
            Path:        r.URL.Path,
            RemoteAddr:  r.RemoteAddr,
            UserAgent:   r.UserAgent(),
            Status:      ww.Status(),
            Duration:    duration,
            TimeStamp:   time.Now(),
        }


        logMessage := fmt.Sprintf(
            "[%s] %s %s %d %v %s %s",
            entry.TimeStamp.Format("2006-01-02 15:04:05"),
            entry.Method,
            entry.Path,
            entry.Status,
            entry.Duration,
            entry.RemoteAddr,
            entry.UserAgent,
        )

       
        logger.Println(logMessage)
        fmt.Println(logMessage)
    })
}