package main

import (
	"testing"
	"github.com/LeBronQ/kdtree"
)

func Benchmark_main(b *testing.B) {
	/*
	for n := 0; n < b.N; n++ {
		main()
	}*/
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		var nodes []kdtree.Point
		NodeArr := make([]*Node, NodeNum)
		NodeArr = GenerateNodes()
		nodes = UpdatePosition(NodeArr, nodes)
		
		tree := kdtree.New(nodes)
		b.StartTimer()
		var graph [NodeNum][]*Node
		UpdateNeighborsAndCalculatePLR(graph, NodeArr, tree)
		
	}
	
}
