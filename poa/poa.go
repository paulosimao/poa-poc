package poa

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

const CycleDuration = 1
const ORIGIN = 0

// Shared data
var nodes []*Node
var nodeCount = 0
var cycleCount = 0
var lastHashVal = 0
var newHashVal = 0
var chSync chan struct{}
var chChange chan struct{}
var lastWinner *Node

// removeNode removes item at position i in our array.
func removeNode(nodes []*Node, i int) []*Node {
	if i < 0 || i >= len(nodes) {
		return nodes
	}
	nodes[i] = nodes[len(nodes)-1]
	nodes = nodes[:len(nodes)-1]
	return nodes
}

// nodes2String simplifies seeing the nodes in a log.
func nodes2String() string {
	sb := strings.Builder{}
	for i := range nodes {
		sb.WriteString(fmt.Sprintf("%v[%v] ", nodes[i].id, i))
	}
	if lastWinner != nil {
		sb.WriteString(fmt.Sprintf("=== LW:%v[%v] ", lastWinner.id, "LW"))
	}
	sb.WriteString(fmt.Sprintf("  LEN:%v", len(nodes)))
	return sb.String()
}

// Origin returns the origin node.
func Origin() *Node {
	return nodes[ORIGIN]
}

// =============================================================================

// Node represents the node struct.
type Node struct {
	keepRunning bool
	id          int
	denial      bool
}

// idx - finds the index of item i among de nodes.
func (n *Node) idx() int {
	for i, nn := range nodes {
		if nn == n {
			return i
		}
	}
	// Dude, where did you go? This node is no longer in the nodes pool.
	// This may happen if it was the last winner, and getting some vacations
	// during the next cycle, or... it may have produced a wrong result and
	// was slashed by origin.
	return -1
}

// Start will spin up the node.
func (n *Node) Start() {

	// In case of nodes rejoining, for tracing purposes we would rather stick
	// with their original ids.
	if n.id == 0 {
		nodeCount++
		n.id = nodeCount
	}
	n.keepRunning = true

	nodes = append(nodes, n)

	//If this is the first node we will setup the whole env.
	if n == Origin() {
		chSync = make(chan struct{})
		chChange = make(chan struct{})
		rand.Seed(time.Now().UnixNano())
		lastHashVal = rand.Int()
	}

	log.Printf("Node %v[%v] starting. Origin now: %v", n.id, n.idx(), Origin().id)
	go n.loop()
}

// stop stops this node.
func (n *Node) stop() {
	n.keepRunning = false

	// This go routine allows a node to rejoin the pool after a while.
	go func() {
		time.Sleep(time.Second * 5)
		n.Start()
	}()
}

// loop will call cycle over and over again.
func (n *Node) loop() {
	for n.keepRunning {
		n.cycle()
	}
}

// startCycle is called when we need to prepare a new cycle.
func startCycle() {
	cycleCount++
	log.Printf("Origin: %v[%v] - nodes: %s", Origin().id, Origin().idx(), nodes2String())
	lastHashVal = newHashVal
	close(chSync)
	chSync = make(chan struct{})
}

// syncStartCycle will ensure all nodes are in sync and also that in case
// a new origin.
// This design needs improvements.
// takes over this is properly
// Getting deadlocks at high rate (50ms for mining.)
// Above 1 sec everything seems to be fine.
// Also need to double check how is it behaving when we extinguish the pool.
func (n *Node) syncStartCycle() {

	if n == Origin() {
		startCycle()
	} else {
		select {
		case <-chChange: // Happens when a new Origin is selected.
			if n == Origin() {
				startCycle()
			} else {
				<-chSync
			}
		case <-chSync:
		}

	}

}

// closeCycle is called at the end of the cycle, by Origin and does the closure
// of the computation performed.
func (n *Node) closeCycle(actWinnerId int) {

	if n == Origin() {
		// This is where we give space for other nodes to do their computational
		// tasks. We wait for the cycle here.
		<-time.After(CycleDuration * time.Millisecond * 1000)

		actWinner := nodes[actWinnerId]

		// For simulating failures, we perceive multiples of 9 as errors.
		if (lastHashVal == newHashVal || lastHashVal == -1 || newHashVal%9 == 0) && actWinner != Origin() && len(nodes) > 1 {
			log.Printf("****Node %v[%v] failed mining - removing from pool.", actWinner.id, actWinnerId)
			log.Printf("%v => %v", lastHashVal, newHashVal)
			actWinner.stop()
			actWinner = nil
		}

		if len(nodes) > 1 {
			nodes = removeNode(nodes, actWinnerId)
		}

		if lastWinner != nil {
			nodes = append(nodes, lastWinner)
		}
		lastWinner = actWinner

		close(chChange)
		chChange = make(chan struct{})
	}

}

// cycle represents one mining cycle.
func (n *Node) cycle() bool {

	// Lets ensure all nodes are in sync for the start of a new cycle.
	n.syncStartCycle()

	// This is a very simplified version of election algorithm.
	actWinnerId := lastHashVal % len(nodes)

	// Lets go mining?
	if actWinnerId == n.idx() {
		log.Printf("  Winner to mine: %v[%v]", n.id, n.idx())
		n.mine(cycleCount, lastHashVal)
	}
	// Close this cycle - mostly relevant to origin
	n.closeCycle(actWinnerId)

	return true
}

// mine simulates mining by providing a random number.
func (n *Node) mine(cycleCount int, actHash int) {
	nh := rand.Int()
	log.Printf("    Node: %v[%v] mined: %v => Previous: %v", n.id, n.idx(), nh, actHash)
	newHashVal = nh
}
