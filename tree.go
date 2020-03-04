package main

import (
	"fmt"
)

type Node struct {
	value    string
	children map[*Node]struct{}
}

var tree Node

func initTree() {
	tree.value = "/"
	tree.children = make(map[*Node]struct{})
}

func main() {

	initTree()
	node1 := Node{value: "1"}
	node1.children = make(map[*Node]struct{})

	node2 := Node{value: "2"}
	node2.children = make(map[*Node]struct{})

	node3 := Node{value: "3"}
	node3.children = make(map[*Node]struct{})

	node21 := Node{value: "2-1"}
	node21.children = make(map[*Node]struct{})

	node22 := Node{value: "2-2"}
	node22.children = make(map[*Node]struct{})

	node31 := Node{value: "3-1"}
	node31.children = make(map[*Node]struct{})

	// agregar
	tree.children[&node1] = struct{}{}
	tree.children[&node2] = struct{}{}
	tree.children[&node3] = struct{}{}

	node2.children[&node21] = struct{}{}
	node2.children[&node22] = struct{}{}

	node3.children[&node31] = struct{}{}

	print()
}

func print() {
	var recPath func(level int, node Node)
	recPath = func(level int, node Node) {
		for i := 0; i < level; i++ {
			fmt.Print("|  ")
		}
		fmt.Println(node.value)
		if node.children != nil && len(node.children) > 0 {
			for pointer := range node.children {
				recPath(level+1, *pointer)
			}
		}
	}

	recPath(0, tree)
}
