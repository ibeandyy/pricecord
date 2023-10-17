package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"pricecord/pkg/Controller"
	"syscall"
)

func main() {
	token := flag.String("token", "", "Discord Bot Token")
	flag.Parse()

	if *token == "" {
		fmt.Println("Please provide a token")
		os.Exit(1)
	}

	c := controller.NewController(*token)

	go c.Initialize()
	defer c.Close()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals

	fmt.Printf("Received signal: %v. Shutting down...\n", sig)

	os.Exit(0)

}
