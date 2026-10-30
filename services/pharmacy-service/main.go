package main

import (
"context"
"encoding/json"
"log"
"net"
"net/http"
"os"
"os/signal"
"syscall"
"time"
)

var stock = map[string]int{"MED-AMX-500": 120, "MED-PCT-500": 300}

func env(k, d string) string {
if v := os.Getenv(k); v != "" {
return v
}
return d
}

func tcpUp(addr string) bool {
c, err := net.DialTimeout("tcp", addr, 2*time.Second)
if err != nil {
return false
}
c.Close()
return true
}

func watch(ctx context.Context, name, addr string) {
was, first := false, true
for {
up := tcpUp(addr)
if up != was || first {
if up {
log.Printf("%s UP (%s)", name, addr)
} else {
log.Printf("%s DOWN (%s), retrying", name, addr)
}
was, first = up, false
}
select {
case <-ctx.Done():
return
case <-time.After(3 * time.Second):
}
}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(code)
json.NewEncoder(w).Encode(v)
}

func main() {
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

mongo := env("MONGO_ADDR", "pharmacy-db:27017")
rabbit := env("RABBITMQ_ADDR", "message-broker:5672")
go watch(ctx, "mongodb", mongo)
go watch(ctx, "rabbitmq", rabbit)

// Placeholder port gRPC (versi sederhana: hanya listener TCP)
if ln, err := net.Listen("tcp", ":"+env("GRPC_PORT", "50051")); err == nil {
log.Printf("grpc placeholder listening on :%s", env("GRPC_PORT", "50051"))
go func() {
<-ctx.Done()
ln.Close()
}()
go func() {
for {
c, err := ln.Accept()
if err != nil {
return
}
c.Close()
}
}()
}

mux := http.NewServeMux()
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
if !tcpUp(mongo) {
writeJSON(w, 503, map[string]string{"service": "pharmacy-service", "status": "degraded", "db": "down"})
return
}
writeJSON(w, 200, map[string]string{"service": "pharmacy-service", "status": "ok", "db": "up"})
})
mux.HandleFunc("/api/v1/drugs/check", func(w http.ResponseWriter, r *http.Request) {
code := r.URL.Query().Get("drug_code")
qty, ok := stock[code]
if !ok {
writeJSON(w, 404, map[string]string{"error": "drug not found", "drug_code": code})
return
}
writeJSON(w, 200, map[string]any{"drug_code": code, "stock": qty, "available": qty > 0})
})

srv := &http.Server{Addr: ":" + env("REST_PORT", "8081"), Handler: mux}
go func() {
<-ctx.Done()
log.Println("SIGTERM received, shutting down gracefully")
sctx, c := context.WithTimeout(context.Background(), 5*time.Second)
defer c()
srv.Shutdown(sctx)
}()
log.Printf("pharmacy-service REST listening on %s", srv.Addr)
if err := srv.ListenAndServe(); err != http.ErrServerClosed {
log.Fatal(err)
}
}
