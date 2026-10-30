package main

import (
"context"
"database/sql"
"encoding/json"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

_ "github.com/lib/pq"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(code)
json.NewEncoder(w).Encode(v)
}

func main() {
port := os.Getenv("PORT")
if port == "" {
port = "8080"
}
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
log.Fatalf("invalid DATABASE_URL: %v", err)
}
db.SetConnMaxLifetime(30 * time.Second)

ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

// Monitor koneksi DB: log hanya saat status berubah (untuk uji self-healing)
go func() {
up, first := false, true
for {
cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
err := db.PingContext(cctx)
cancel()
if (err == nil) != up || first {
if err == nil {
log.Println("database UP: connection established")
} else {
log.Printf("database DOWN: %v (retrying)", err)
}
up, first = err == nil, false
}
select {
case <-ctx.Done():
return
case <-time.After(3 * time.Second):
}
}
}()

mux := http.NewServeMux()
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
if err := db.PingContext(r.Context()); err != nil {
writeJSON(w, 503, map[string]string{"service": "appointment-service", "status": "degraded", "db": "down"})
return
}
writeJSON(w, 200, map[string]string{"service": "appointment-service", "status": "ok", "db": "up"})
})
mux.HandleFunc("/api/v1/appointments", func(w http.ResponseWriter, r *http.Request) {
var n int
if err := db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM appointments").Scan(&n); err != nil {
writeJSON(w, 503, map[string]string{"error": "database unavailable"})
return
}
writeJSON(w, 200, map[string]int{"total_appointments": n})
})

srv := &http.Server{Addr: ":" + port, Handler: mux}
go func() {
<-ctx.Done()
log.Println("SIGTERM received, shutting down gracefully")
sctx, c := context.WithTimeout(context.Background(), 5*time.Second)
defer c()
srv.Shutdown(sctx)
}()
log.Printf("appointment-service listening on :%s", port)
if err := srv.ListenAndServe(); err != http.ErrServerClosed {
log.Fatal(err)
}
}
