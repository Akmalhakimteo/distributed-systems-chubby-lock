package main

import (
	"fmt"
	"time"
	"os"
	"strconv"
)

type Message struct{
	fileName string
	fileContent string
}

type Node struct{
	id int
	revChan chan Message
	replyChan chan bool
	peers []Node
	quitElect chan int
	killNode chan int
	coordinator int
}

func main() {
	arg := os.Args[1] //get the number of nodes
	numNodes, _ := strconv.Atoi(arg) //set this as the number of nodes we have

	
	allNodes := make([]Node, numNodes) //make the nodes

	for i:= 0; i< numNodes; i++{
		allNodes[i].id = i
		allNodes[i].revChan = make(chan Message) //from master server to the server
		allNodes[i].replyChan = make(chan bool) //from servers to master
		// allNodes[i].quitElect = make(chan int)
		// allNodes[i].killNode = make(chan int)
		// allNodes[i].peers = allNodes
		// allNodes[i].coordinator = numNodes-1 
	}
	//channels between master server and client
	fromClient := make(chan Message)
	toClient := make(chan bool)

	go clientWriteReq("new file content!", "FileABC", fromClient)
	go masterPropagate(fromClient, allNodes, toClient)

	for i:= 0; i< len(allNodes)-1; i++{
		go cellServerOK(i+1, allNodes[i].revChan, allNodes[i].replyChan)
	}

	go cellServerFail(4, allNodes[len(allNodes)-1].revChan, allNodes[len(allNodes)-1].replyChan)
	go clientAwaitACK(toClient)

	//take grace period as 15 seconds, if not reponse close program
	time.Sleep(15 * time.Second)
	fmt.Println("program closing...")
}

func clientWriteReq(fileContent string, fileName string, fromClient chan Message) {
	fmt.Println("client request to write to", fileName, "with content:", fileContent)
	data := Message{fileName, fileContent}
	fromClient <- data
}

func masterPropagate(fromClient chan Message, allNodes []Node, toClient chan bool) {
	for {
		data := <-fromClient
		fileName := data.fileName
		fileContent := data.fileContent
		fmt.Println("server has received request to write to", fileName)
		for i:=0; i< len(allNodes); i++{
			allNodes[i].revChan <- data
		}
		fmt.Println("server has propagated request to write to entire cell, awaiting response...")

		//receiving responses from servers in cell
		//master wait for servers to repsond w true, else will not continue

		OK := true
		for i:=0; i<len(allNodes); i++{
			checkBool := <- allNodes[i].replyChan
			if checkBool != true{
				fmt.Println("error msg from server ", i+ 1, "\n")
				OK = false
			}

		}
		
		if OK {
			fmt.Println("master has received ACK from all servers")
			//master updating database
			time.Sleep(5 * time.Second)
			fmt.Println("master finished updating", fileName, "w", fileContent)
		} else {
			fmt.Println("write failure on replica server, master refused update")
		}
		toClient <- OK
	}
}

func cellServerOK(index int, fromMaster chan Message, toMaster chan bool) {
	data := <-fromMaster
	fileName := data.fileName
	fileContent := data.fileContent
	// server updates replicas
	time.Sleep(5 * time.Second)
	fmt.Println("server", index, "finished updating", fileName, "w", fileContent)
	toMaster <- true
}

func cellServerFail(index int, fromMaster chan Message, toMaster chan bool) {
	data := <-fromMaster
	fileName := data.fileName
	//fileContent := data.fileContent
	// server updates replicas
	time.Sleep(5 * time.Second)
	fmt.Println("server", index, "error updating", fileName)
	toMaster <- false
}

func clientAwaitACK(toClient chan bool) {
	for {
		ACK := <-toClient
		if ACK {
			fmt.Println("SUCCESS")
		} else {
			fmt.Println("FAIL :(")
		}
	}
}
