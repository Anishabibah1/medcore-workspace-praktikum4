package main

import (
"context"
"log"
"net"
"os"
"os/signal"
"syscall"
"time"
)

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

func main() {
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()
log.Println("billing-service worker started")
go watch(ctx, "mysql", env("MYSQL_ADDR", "billing-db:3306"))
go watch(ctx, "rabbitmq", env("RABBITMQ_ADDR", "message-broker:5672"))
<-ctx.Done()
log.Println("SIGTERM received, shutting down gracefully")
}
