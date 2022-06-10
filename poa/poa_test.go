package poa_test

import (
	"poa/poa"
	"testing"
	"time"
)

func TestNodes(t *testing.T) {
	//var nodes []*poa.Node
	for i := 0; i < 10; i++ {
		var n poa.Node
		n.Start()
		//nodes = append(nodes, n)
	}
	time.Sleep(poa.CycleDuration * 5 * time.Hour)
}
