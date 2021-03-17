package main

import (
	"fmt"
	"time"
	// "os"
	// "strconv"
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
	// arg := os.Args[1] //get the number of nodes
	// numNodes, _ := strconv.Atoi(arg) //set this as the number of nodes we have

	
	// allNodes := make([]Node, numNodes) //make the nodes

	// for i:= 0; i< numNodes; i++{
	// 	allNodes[i].id = i
	// 	allNodes[i].revChan = make(chan Message) //from master server to the server
	// 	allNodes[i].replyChan = make(chan bool) //from servers to master
	// 	allNodes[i].quitElect = make(chan int)
	// 	allNodes[i].killNode = make(chan int)
	// 	allNodes[i].peers = allNodes
	// 	allNodes[i].coordinator = numNodes-1 
	// }
	//channels between master server and client
	fromClient := make(chan Message)
	toClient := make(chan bool)
	//channels from master server [0] to other servers in cell [1-4]
	M1 := make(chan Message)
	M2 := make(chan Message)
	M3 := make(chan Message)
	M4 := make(chan Message)
	//channels from servers in cell [1-4] to master server [0]
	S1 := make(chan bool)
	S2 := make(chan bool)
	S3 := make(chan bool)
	S4 := make(chan bool)

	go clientWriteReq("new file content!", "FileABC", fromClient)
	go masterPropagate(fromClient, M1, M2, M3, M4, S1, S2, S3, S4, toClient)
	go cellServerOK(1, M1, S1)
	go cellServerOK(2, M2, S2)
	go cellServerOK(3, M3, S3)
	//go cellServerOK(4, M4, S4)
	go cellServerFail(4, M4, S4)
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

func masterPropagate(fromClient chan Message, M1 chan Message, M2 chan Message, M3 chan Message, M4 chan Message, S1 chan bool, S2 chan bool, S3 chan bool, S4 chan bool, toClient chan bool) {
	for {
		data := <-fromClient
		fileName := data.fileName
		fileContent := data.fileContent
		fmt.Println("server has received request to write to", fileName)
		M1 <- data
		M2 <- data
		M3 <- data
		M4 <- data
		fmt.Println("server has propagated request to write to entire cell, awaiting response...")

		//receiving responses from servers in cell
		//master wait for servers to repsond w true, else will not continue

		ACK1 := <-S1
		ACK2 := <-S2
		ACK3 := <-S3
		ACK4 := <-S4

		OK := ACK1 && ACK2 && ACK3 && ACK4 //if all ACKS are received, OK == 1 

		if !ACK1 {
			fmt.Println("error msg from server 1")
		}
		if !ACK2 {
			fmt.Println("error msg from server 2")
		}
		if !ACK3 {
			fmt.Println("error msg from server 3")
		}
		if !ACK4 {
			fmt.Println("error msg from server 4")
		}

		//master received all ACKS
		fmt.Println("master has received ACK from all servers")
		if OK {
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
