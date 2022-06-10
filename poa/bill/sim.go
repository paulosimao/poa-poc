package main

import (
	"log"
	"sync"
	"time"
)

// block represents a block of data that is mined.
type block struct{}

// node represents a node running as a process.
type node struct {
	name   string
	wg     sync.WaitGroup
	ticker *time.Ticker
	shut   chan struct{}
	send   chan block
}

// run gets the node up and running.
func run(name string) *node {
	n := node{
		name:   name,
		ticker: time.NewTicker(2 * time.Second),
		shut:   make(chan struct{}),
		send:   make(chan block, 10),
	}

	n.wg.Add(1)

	go func() {
		defer n.wg.Done()

		log.Println(n.name, ":starting node")

		for {
			select {
			case <-n.ticker.C:
				n.performWork()
			case <-n.shut:
				log.Println(n.name, ":node shutting down")
				return
			}
		}
	}()

	return &n
}

// shutdown terminates the node from existence.
func (n *node) shutdown() {
	close(n.shut)
	n.wg.Wait()
}

// performWork represents the work to perform on each 12 second cycle.
func (n *node) performWork() {
	log.Println(n.name, ":performing work")
}

// =============================================================================

// simulation represents a set of nodes talking to each other.
type simulation struct {
	nodes []*node
}

// newSimulation starts 3 nodes for the simulation.
func newSimulation() *simulation {
	nodes := []*node{
		run("nodeA"),
		run("nodeB"),
		run("nodeC"),
	}

	return &simulation{
		nodes: nodes,
	}
}

// shutdown turns all the nodes off.
func (s *simulation) shutdown() {
	for _, n := range s.nodes {
		n.shutdown()
	}
}
