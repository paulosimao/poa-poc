package main

import (
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// block represents a block of data that is mined.
type block struct{}

// node represents a node running as a process.
type node struct {
	name     string
	wg       sync.WaitGroup
	ticker   *time.Ticker
	shut     chan struct{}
	send     chan block
	registry map[string]*node
}

// newNode gets the node up and running.
func newNode(name string) *node {
	n := node{
		name:   name,
		ticker: time.NewTicker(2 * time.Second),
		shut:   make(chan struct{}),
		send:   make(chan block, 10),
	}

	return &n
}

// run gets the node up and running.
func (n *node) run(registry map[string]*node) {
	n.registry = registry

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
}

// shutdown terminates the node from existence.
func (n *node) shutdown() {
	close(n.shut)
	n.wg.Wait()
}

// performWork represents the work to perform on each 12 second cycle.
func (n *node) performWork() {
	selectedNode := n.selection()
	switch selectedNode {
	case n.name:
		log.Println(n.name, ":SELECTED")
	default:
	}
}

// selection selects an index from 0 to 2.
func (n *node) selection() string {

	// HOW DO WE MAKE THIS DETERMINISTIC!!!!

	// Sort the current list of registered nodes.
	nodes := make([]string, 0, len(n.registry))
	for key := range n.registry {
		nodes = append(nodes, key)
	}
	sort.Strings(nodes)

	// Return the name of the node selected.
	return nodes[rand.Intn(len(nodes))]
}

// =============================================================================

// simulation represents a set of nodes talking to each other.
type simulation struct {
	nodes map[string]*node
}

// newSimulation starts 3 nodes for the simulation.
func newSimulation() *simulation {
	nodes := map[string]*node{
		"nodeA": newNode("nodeA"),
		"nodeB": newNode("nodeB"),
		"nodeC": newNode("nodeC"),
	}

	for _, n := range nodes {
		n.run(nodes)
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
