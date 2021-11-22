package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"


	"github.com/Mateus-pilo/go-whats-opt/hlp"
	"github.com/Mateus-pilo/go-whats-opt/hlp/libs"
	"github.com/Mateus-pilo/go-whats-opt/hlp/router"
	
)

// Server Variable
var svr *hlp.Server

// Init Function
func init() {
	// Initialize Server
	svr = hlp.NewServer(router.Router)
}

// Main Function
func main() {
	// Starting Server
	_ = libs.ConnectionMqp();
	svr.Start()

	// Make Channel for OS Signal
	sig := make(chan os.Signal, 1)

	// Notify Any Signal to OS Signal Channel
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	// Return OS Signal Channel
	// As Exit Sign
	<-sig

	// Log Break Line
	fmt.Println("")

	// Stopping Server
	defer svr.Stop()
}
