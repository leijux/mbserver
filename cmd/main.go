package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/leijux/mbserver"
)

var addr = flag.String("addr", ":8080", "TCP address to listen on")

func main() {
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	s := mbserver.NewServer()
	s.ListenTCP(*addr)

	defer s.Shutdown()

	go s.Start()

	<-sigs
}
