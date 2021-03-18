package main

import (
	s "ds_proj/server"
	"os"
	"strconv"
	"fmt"
	c "ds_proj/client"
	// "time"
)


func main(){
	var (
		allNodes []s.Node
		allClients []c.Client
		allChans []chan c.SelfIntroduction
	)
	arg := os.Args[1]
	numNodes, _ := strconv.Atoi(arg)
	allNodes = s.Start(numNodes)

	for i:= 0; i< numNodes; i++{
		allChans = append(allChans, allNodes[i].ClientChan)
	}

	// time.Sleep(time.Second * 20)
	allClients = c.InitClient(3, allChans)

	fmt.Println(allNodes[0].Coordinator)
	fmt.Println(allClients[0].Id)
	// s.KillNode(4, allNodes)
	// s.SimulateDuplicateACK()
	// s.WriteToAllNodes(allNodes,numNodes,"newfile.txt","newcontent")
	var input string
	fmt.Scanln(&input)
}