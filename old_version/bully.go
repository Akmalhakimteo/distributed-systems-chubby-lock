package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Node struct{
	id int
	revChan chan Message
	replyChan chan Message
	peers []Node
	quitElect chan int
	killNode chan int
	coordinator int
}

type Message struct{
	senderID int
	msg string
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

func (n *Node) elect(){
	go n.checkReply()
	for i :=n.id+1; i<len(n.peers); i++{
		fmt.Println("Send Elect to: ", i)
		send(n.peers[i].revChan, Message{n.id, "Elect"})
		//n.peers[i].revChan <- Message{n.id, "Elect"}
		<- n.quitElect
	}
}


func (n *Node) checkChannel(){
	for {
		select{
		case x := <-n.revChan:
			if x.msg == "Elect"{
				// reply if id > sender's id
				fmt.Println(n.id, "Received ELECT")
				if x.senderID < n.id{
					// Send reply to sender to challenge election
					fmt.Println(n.id, "Challenging election")
					send(n.peers[x.senderID].replyChan, Message{n.id, "Reply"})
					go n.elect()
				}
			}else if x.msg == "Coordinate"{
				n.coordinator = x.senderID
				fmt.Println(n.id, "Received COORDINATOR: ", x.senderID)
			}else if x.msg == "BLOCKED"{
				return
			}
		case <- n.killNode:
			return
		}
	}
}

func (n *Node) checkReply(){
	noReply := 0
	for i:= 0; i<5; i++{
		time.Sleep(time.Millisecond * 500)
		select{
		case <- n.replyChan:
			// stop election process
			fmt.Println(n.id, "Received REPLY")
			sendInt(n.quitElect, 0)

		default:
			//nothing receive
			fmt.Println("No reply received")
			noReply ++
			continue
		}
	}
	if noReply == 5{
		fmt.Println("MUAHAHA IM THE BULLY NOW")
		n.coordinator = n.id
		for i:= 0; i<n.id; i++{
			send(n.peers[i].revChan, Message{n.id, "Coordinate"})
		}
	}
}

func (n *Node) kill(){
	sendInt(n.killNode, 0)
	//n.killNode <- 0
}


func main(){
	 arg := os.Args[1]
	 numNodes, _ := strconv.Atoi(arg)

	allNodes := make([]Node, numNodes)

	for i:= 0; i< numNodes; i++{
		allNodes[i].id = i
		allNodes[i].revChan = make(chan Message)
		allNodes[i].replyChan = make(chan Message)
		allNodes[i].quitElect = make(chan int)
		allNodes[i].killNode = make(chan int)
		allNodes[i].peers = allNodes
		allNodes[i].coordinator = numNodes-1
	}
	toKill := -1

	//Uncomment line 149 to test best case of election
	go allNodes[0].elect()

	//Uncomment line 152 to test worst case of election
	// go allNodes[numNodes-2].elect()

	//Uncomment lines 155-157 to test random killing of nodes
	toKill = rand.Intn(numNodes)
	fmt.Println("We're Killing node: ", toKill)
	allNodes[toKill].kill()


	for i:=0; i<numNodes; i++{
		go allNodes[i].checkChannel()
	}



	//var input string
	//fmt.Scanln(&input)
	time.Sleep(time.Second * 7)
	for i:=0; i<numNodes; i++{
		if i!= toKill{
			fmt.Println("Node",  i, "'s coordinator is: ", allNodes[i].coordinator)
		}
	}
}