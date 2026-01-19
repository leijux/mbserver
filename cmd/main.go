package main

import (
	"context"
	"flag"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/leijux/mbserver"
)

var addr = flag.String("addr", ":8080", "TCP address to listen on")

func main() {
	flag.Parse()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	s := mbserver.NewServer()

	err := s.ListenTCP(*addr)
	if err != nil {
		slog.Error("listen tcp err", "err", err)
		return
	}

	defer s.Shutdown()

	go s.Start()

	<-ctx.Done()
}
