package main

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// block represents a block of data that is mined.
type block struct {
	Number        uint64
	PrevBlockHash string
	TimeStamp     uint64
}

// newBlock constructs a block from the previous block.
func newBlock(prevBlock block) block {
	return block{
		Number:        prevBlock.Number + 1,
		PrevBlockHash: prevBlock.hash(),
		TimeStamp:     uint64(time.Now().UTC().UnixMilli()),
	}
}

// hash generates the hash for this block.
func (b block) hash() string {
	const ZeroHash string = "0x0000000000000000000000000000000000000000000000000000000000000000"

	data, err := json.Marshal(b)
	if err != nil {
		return ZeroHash
	}

	hash := sha256.Sum256(data)
	return hexutil.Encode(hash[:])
}

// =============================================================================

// node represents a node running as a process.
type node struct {
	name        string
	wg          sync.WaitGroup
	ticker      *time.Ticker
	shut        chan struct{}
	send        chan block
	registry    map[string]*node
	latestBlock block
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
			case b := <-n.send:
				log.Println(n.name, ":node received block", b.hash())
				n.latestBlock = b
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
		n.mineNewBlock()
	default:
	}
}

// selection selects an index from 0 to 2.
func (n *node) selection() string {

	// Sort the current list of registered nodes.
	nodes := make([]string, 0, len(n.registry))
	for key := range n.registry {
		nodes = append(nodes, key)
	}
	sort.Strings(nodes)

	// HOW DO WE MAKE THIS DETERMINISTIC!!!!
	// FOR NOW JUST USE INDEX 0

	// Return the name of the node selected.
	return nodes[0]
}

// mineNewBlock creates a new block and sends that to the p2p network.
func (n *node) mineNewBlock() {
	b := newBlock(n.latestBlock)
	n.latestBlock = b

	// Send block to all other nodes in the registry.
	for k, node := range n.registry {
		if k != n.name {
			node.send <- b
		}
	}
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
