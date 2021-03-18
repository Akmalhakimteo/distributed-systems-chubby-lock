package server

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	// "math/rand"
	c "ds_proj/client"
)

type Node struct{
	id int
	revChan chan Message
	replyChan chan Message
	ClientChan chan c.SelfIntroduction
	peers []Node
	quitElect chan int
	killNode chan int
	Coordinator int
	AllClients []chan c.File
}

type Message struct{
	SenderID int
	Msg string
}

type Lock struct {
	filename string
	mode string
	owners map[int]bool
}

func Start(numNodes int) []Node {
	allNodes := make([]Node, numNodes)

	for i := 0; i < numNodes; i++ {
		allNodes[i].id = i
		allNodes[i].revChan = make(chan Message)
		allNodes[i].replyChan = make(chan Message)
		allNodes[i].quitElect = make(chan int)
		allNodes[i].killNode = make(chan int)
		allNodes[i].peers = allNodes
		allNodes[i].Coordinator = numNodes - 1
		allNodes[i].ClientChan = make(chan c.SelfIntroduction)
	}

	//Uncomment line 149 to test best case of election
	// go allNodes[0].elect()

	for i := 0; i < numNodes; i++ {
		go allNodes[i].ping()
		go allNodes[i].checkChannel()
	}

	// var input string
	// fmt.Scanln(&input)
	// time.Sleep(time.Second * 3)
	// for i:=0; i<numNodes; i++{
	// 	fmt.Println("Node",  i, "'s coordinator is: ", allNodes[i].coordinator)
	// }
	return allNodes
}

func KillNode(id int, allNodes []Node) {
	// id = rand.Intn(numNodes)
	fmt.Println("We're Killing node: ", id)
	allNodes[id].kill()
	// time.Sleep(time.Second * 3)
	// for i:=0; i<len(allNodes); i++{
	// 	fmt.Println("Node",  i, "'s coordinator is: ", allNodes[i].coordinator)
	// }
}

func timer(d time.Duration) chan bool {
	ch := make(chan bool, 1)
	go func() {
		time.Sleep(d)
		ch <- true
	}()
	return ch
}

func send(ch chan Message, msg Message) chan bool {
	outCh := make(chan bool, 1)
	go func() {
		t := timer(time.Duration(3 * time.Second))
		select {
		case <-t:
			outCh <- false
		case ch <- msg:
			outCh <- true
		}
	}()
	return outCh
}

func sendInt(ch chan int, msg int) chan bool {
	outCh := make(chan bool, 1)
	go func() {
		t := timer(time.Duration(3 * time.Second))
		select {
		case <-t:
			outCh <- false
		case ch <- msg:
			outCh <- true
		}
	}()
	return outCh
}

func (n *Node) elect() {
	go n.checkReply()
	for i := n.id + 1; i < len(n.peers); i++ {
		fmt.Println("Send Elect to: ", i)
		send(n.peers[i].revChan, Message{n.id, "Elect"})
		//n.peers[i].revChan <- Message{n.id, "Elect"}
		<-n.quitElect
	}
}

func (n *Node) checkChannel() {
	for {
		select {
		case x := <-n.revChan:
			if x.Msg == "Elect"{
				// reply if id > sender's id
				fmt.Println(n.id, "Received ELECT")
				if x.SenderID < n.id{
					// Send reply to sender to challenge election
					fmt.Println(n.id, "Challenging election")
					send(n.peers[x.SenderID].replyChan, Message{n.id, "Reply"})
					go n.elect()
				}
			}else if x.Msg == "Coordinate"{
				n.Coordinator = x.SenderID
				fmt.Println(n.id, "Received COORDINATOR: ", x.SenderID)
			}else if x.Msg == "BLOCKED"{
				return
			}
		case intro := <-n.ClientChan:
			if intro.Msg == "req connection"{
				if n.id!=n.Coordinator{
					continue
				}
				select{
				case intro.RevChan<-c.File{len(n.AllClients), "", ""}:
					fmt.Println("connected with client")
					n.AllClients = append(n.AllClients, intro.RevChan)
					fmt.Println("current length of clients list: ", len(n.AllClients))
				case <- time.After(time.Second*5):
					fmt.Println("failed to connect")
				}
			}else if intro.Msg == "req master"{
				select{
				case intro.RevChan<-c.File{len(n.AllClients), "", strconv.Itoa(n.Coordinator)}:
					fmt.Println("sent client the coordinator")
				case <- time.After(time.Second*5):
					fmt.Println("failed to send client the coordinator")
				}
			}
		case <- n.killNode:
			return
		}
	}
}

func (n *Node) ping() {
	for {
		// fmt.Println("peers: ", n.peers)
		fmt.Println("node: ", n.id, "is pinging master server: ", n.Coordinator)
		select{
		case n.peers[n.Coordinator].revChan<-Message{n.id, "Are you alive?"}:
			time.Sleep(time.Second * 3)
		case <-time.After(time.Second * 10):
			n.elect()
		case <-n.killNode:
			return
		}
	}
}

func (n *Node) checkReply() {
	noReply := 0
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 1000)
		select {
		case <-n.replyChan:
			// stop election process
			fmt.Println(n.id, "Received REPLY")
			sendInt(n.quitElect, 0)

		default:
			//nothing receive
			fmt.Println("No reply received")
			noReply++
			continue
		}
	}
	if noReply == 5 {
		fmt.Println("MUAHAHA IM THE BULLY NOW")
		n.Coordinator = n.id
		for i:= 0; i<n.id; i++{
			send(n.peers[i].revChan, Message{n.id, "Coordinate"})
		}
	}
}

func (n *Node) kill() {
	sendInt(n.killNode, 0)
	sendInt(n.killNode, 0)
	//n.killNode <- 0
}

func SimulateDuplicateACK() {
	fmt.Println("simulating duplicate protocol")
}

func WriteToAllNodes(allNodes []Node, numNodes int, fileName, fileContent string) {
	for i := 0; i < numNodes; i++ {
		nodeNumFile := "node" + strconv.Itoa(allNodes[i].id)
		nodeNumFile += ".txt"
		// fmt.Println("Printing node: ",allNodes[i].id)
		// fmt.Println(nodeNumFile,fileContent)
		f, err := os.OpenFile(nodeNumFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		writeToFile := fileName + " " + fileContent + "\n"
		if _, err := f.WriteString(writeToFile); err != nil {
			log.Println(err)
		}
		fmt.Printf("Node %v is updated with most recent file\n", allNodes[i].id)
	}
}

func ServerNodeWriteUpdate(allNodes []Node, numNodes int, fileName, fileContent string) {
	for i := 0; i < numNodes-1; i++ {
		nodeNumFile := "node" + strconv.Itoa(allNodes[i].id)
		nodeNumFile += ".txt"
		// fmt.Println("Printing node: ",allNodes[i].id)
		// fmt.Println(nodeNumFile,fileContent)
		f, err := os.OpenFile(nodeNumFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		writeToFile := fileName + " " + fileContent + "\n"
		if _, err := f.WriteString(writeToFile); err != nil {
			log.Println(err)
		}
		fmt.Printf("Node %v is updated with most recent file\n", allNodes[i].id)
	}
}

func MasterNodeWriteUpdate(masterNodeNumber int, fileName, fileContent string) {
	nodeNumFile := "node" + strconv.Itoa(masterNodeNumber)
	nodeNumFile += ".txt"
	// fmt.Println("Printing node: ",allNodes[i].id)
	// fmt.Println(nodeNumFile,fileContent)
	f, err := os.OpenFile(nodeNumFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	writeToFile := fileName + " " + fileContent + "\n"
	if _, err := f.WriteString(writeToFile); err != nil {
		log.Println(err)
	}
	fmt.Printf("Node %v is updated with most recent file\n", masterNodeNumber)

}
