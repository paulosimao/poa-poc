package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestTicker(t *testing.T) {
	sim := newSimulation()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Println("******* End Simulation *******")
	sim.shutdown()
}
