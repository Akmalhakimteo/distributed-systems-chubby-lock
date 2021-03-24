package client

import (
	"fmt"
	"time"
	// s "ds_proj/server"
	"strconv"
)

type Client struct{
	Id int
	RevChan chan File
	// recvKeepAlive chan KeepAlive
	masterServer int
	AllServers []chan SelfIntroduction
	ServerID int
	MasterChan chan File

	// revChan chan Message
	// replyChan chan Message
	// peers []Node
	// quitElect chan int
	// killNode chan int
	// coordinator int
}

type File struct{
	SenderID int
	Filename string
	Filecontent string
}

type SelfIntroduction struct{
	SenderID int
	RevChan chan File
	Msg string
}

// Function to initialize clients
func InitClient(numClients int, AllServers []chan SelfIntroduction) []Client{
	allClients := make([]Client, numClients)

	for i:= 0; i< numClients; i++{
		allClients[i].Id = i
		allClients[i].RevChan = make(chan File)
		allClients[i].masterServer = -1
		allClients[i].AllServers = AllServers

		// go allClients[j].ClientCheckKeepAlive()
	}

	// Communicate with server to find master and establish connection
	for i:=0; i<numClients; i++{
		go allClients[i].find_master()
		// go allClients[i].WriteReq("test.txt", "Hello World")
	}
	fmt.Println("Clients Created")
	return allClients
}

// Communicates with any of the servers to find the master then connect with the master
func (c *Client) find_master(){
	for i:=0; i<len(c.AllServers); i++{
		select{
		case c.AllServers[i]<-SelfIntroduction{c.Id, c.RevChan, "req master"}:
			fmt.Println("Client: ", c.Id,"Successfully sent message to server: ", i, " now waiting for reply")
			select{
			case reply := <- c.RevChan:
				c.masterServer, _ = strconv.Atoi(reply.Filecontent)
				fmt.Println("Current masterServer is: ", c.masterServer)
				c.connect()
				return
			case <- time.After(time.Second*5):
				fmt.Println("Client: ", c.Id,"failed to connect to server: ", i)
				continue
			}
		case <- time.After(time.Second*5):
			fmt.Println("Client: ", c.Id,"failed to connect to server: ", i)
			continue
		}
	}
	fmt.Println("Failed to connect to any server")
}

// Connect with the master by sending it the client's channel and saving the index that the server saved this client at
func (c *Client) connect(){
	for{
		select{
		case c.AllServers[c.masterServer]<-SelfIntroduction{c.Id, c.RevChan, "req connection"}:
			fmt.Println("Client: ", c.Id,"Successfully sent message to server: ", c.masterServer, " now waiting for reply")
			select{
			case reply := <- c.RevChan:
				c.ServerID = reply.SenderID
				fmt.Println("Successfully connected to current masterServer: ", c.masterServer)
				fmt.Println("ServerID is: ", c.ServerID)
				select{
				case filechan := <- c.AllServers[c.masterServer]:
					c.MasterChan = filechan.RevChan
					fmt.Println("Received file chan from master server")
					return
				case <- time.After(time.Second*5):
					fmt.Println("Client: ", c.Id,"failed to connect to server: ", c.masterServer)
					continue
				}
			case <- time.After(time.Second*5):
				fmt.Println("Client: ", c.Id,"failed to connect to server: ", c.masterServer)
				continue
			}
		case <- time.After(time.Second*5):
			fmt.Println("Client: ", c.Id,"failed to connect to server: ", c.masterServer)
			continue
		}
	}
	fmt.Println("Failed to connect to any server")
}

func (c *Client) WriteReq(fileContent string, fileName string) {
	fmt.Println("client request to write to", fileName, "with content:", fileContent)
	data := File{c.ServerID, fileName, fileContent}
	c.MasterChan <- data
}

func (c *Client) TryAcquireLock(fileName string) {
	fmt.Println("client request to get lock to", fileName)
	data := File{c.ServerID, fileName, "req lock"}
	c.MasterChan <- data
}

func (c *Client) CreateFile(fileName string) {
	fmt.Println("client request to create file: ", fileName)
	data := File{c.ServerID, fileName, "create file"}
	c.MasterChan <- data
}

// func (c *Client) CreateFile(fileContent string, fileName string, fromClient chan Message) {
// 	fmt.Println("client request to write to", fileName, "with content:", fileContent)
// 	data := Message{fileName, fileContent}
// 	fromClient <- data
// }

// func (c *Client) TryAcquireLock(fileContent string, fileName string, fromClient chan Message) {
// 	fmt.Println("client request to write to", fileName, "with content:", fileContent)
// 	data := Message{fileName, fileContent}
// 	fromClient <- data
// }