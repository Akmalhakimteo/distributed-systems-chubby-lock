package main

import (
	s "ds_proj/server"
	"os"
	"strconv"
	"fmt"
)


func main(){
	var (
		allNodes []s.Node
	)
	arg := os.Args[1]
	numNodes, _ := strconv.Atoi(arg)
	allNodes = s.Start(numNodes)
	s.KillNode(4, allNodes)
	s.SimulateDuplicateACK()
	s.WriteToAllNodes(allNodes,numNodes,"newfile.txt","newcontent")
	var input string
	fmt.Scanln(&input)
}