package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"time"
	"os"
)

type Reply struct {
	Data string
}

type CoordReply struct {
	Coord int
	Data  string
}

type Message struct {
	SenderID int
	Msg      string
}

type Client struct {
	id          int
	Coordinator int
	rpcChan     *rpc.Client
	all_ip      [3]string
}

type ClientRequest struct {
	SenderID    int
	Write       int // 1 = write request, 0 = read request
	Filename    []byte
	Filecontent []byte //empty for read request
}

// Client Read Request
// Client should send: filename string
func (client *Client) SendReadRequest(filename []byte) string{
	fmt.Println("Client wants to read file ", string(filename))
	ReadRequest := ClientRequest{SenderID: client.id, Write: 0, Filename: filename, Filecontent: nil}
	var ReadReply Reply
	readrequest_err := client.rpcChan.Call("Listener.GetRequest", ReadRequest, &ReadReply)
	if readrequest_err != nil {
		fmt.Printf("Client Read Request Failed")
	}
	// wait for ack
	// log.Printf(ReadReply.Data)
	return ReadReply.Data
}

// Client Write Request
// Client should send: filename string, value []byte
func (client *Client) SendWriteRequest(filename []byte, filecontent []byte) {
	WriteRequest := ClientRequest{SenderID: client.id, Write: 1, Filename: filename, Filecontent: filecontent}
	var WriteReply Reply
	writerequest_err := client.rpcChan.Call("Listener.GetRequest", WriteRequest, &WriteReply)
	if writerequest_err != nil {
		fmt.Printf("Client Write Request Failed")
	}
	// wait for ack
	// log.Printf(WriteReply.Data)
}

// Client should tryacquirelock() before sending write request
// if tryacquirelock() fails => read() instead
func (client *Client) Write(filename []byte, filecontent []byte) {
	//tryacquirelock()
	fmt.Println("trying to acquire lock")
	clientRequest := ClientRequest{SenderID: client.id, Write: 1, Filename: filename, Filecontent: filecontent}
	var TryAcquireReply Reply
	tryacquire_err := client.rpcChan.Call("Listener.TryAcquire", clientRequest, &TryAcquireReply)
	if tryacquire_err != nil {
		fmt.Println("Try Acquire Lock Failed")
	}

	if TryAcquireReply.Data == "You can have the lock" {
		// write to file
		client.SendWriteRequest(filename, filecontent)
		fmt.Printf("Client writing to file %v with contents %v\n", string(filename), string(filecontent))
	} else if TryAcquireReply.Data == "Someone else has the lock" {
		// sucks to be you, just read the file
		client.SendReadRequest(filename)
		fmt.Println("Client failed write, reading file ", string(filename))
	}
}

func (client *Client) GetCoordinator() {
	var CoordinatorReply CoordReply
	client.rpcChan.Call("Listener.GetCoordinator", client.id, &CoordinatorReply)
	time.Sleep(time.Second * 5)
	log.Printf(CoordinatorReply.Data)
	if CoordinatorReply.Data == "wait" {
		time.Sleep(time.Second * 5)
		client.GetCoordinator()
		return
	}
	newCoordinatorInt := CoordinatorReply.Coord
	if newCoordinatorInt == -1 {
		client.GetCoordinator()
	} else {
		client.Coordinator = newCoordinatorInt
	}
}

func (client *Client) SendKeepAlive(serverInt int) {
	var KeepAliveReply Reply
	keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
	time.Sleep(time.Second * 5)
	log.Printf(KeepAliveReply.Data)
	if keepalive_err != nil {
		fmt.Printf("server node %v died\n", serverInt)
		// get new coordinator
		for ind, curr_ip := range client.all_ip {
			clientChan, err := rpc.Dial("tcp", curr_ip)
			if err != nil {
				fmt.Printf("client connection with server %v error\n", ind)
				continue
			}
			client.rpcChan = clientChan

			client.GetCoordinator()
			clientChan, err = rpc.Dial("tcp", client.all_ip[client.Coordinator])
			if err != nil {
				fmt.Printf("client connection with server %v error\n", client.Coordinator)
			}
			client.rpcChan = clientChan
			client.SendKeepAlive(client.Coordinator)
			break
		}
	} else {
		client.SendKeepAlive(serverInt)
	}
}

func makeClient(id int) *Client {
	all_ip := [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}
	// all_ip := [5]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234", "172.22.0.5:1234", "172.22.0.6:1234"}
	Coordinator := 2
	// Coordinator := 4
	var rpcChan *rpc.Client
	curr_client := Client{id, Coordinator, rpcChan, all_ip}

	return &curr_client
}

func (client *Client) ConsistentRead(filename []byte){
	for{
		content := client.SendReadRequest(filename)
		log.Println("this is the new Client's master:", content)
		time.Sleep(time.Second)
	}
}

func main() {

	//TODO: Client needs to communicate with chubby cell to find out coordinator
	time.Sleep(5 * time.Second)

	id_arg := os.Args[1]
	id, _ := strconv.Atoi(id_arg)
	client := makeClient(id)

	log.Println("Client", id, "is running")

	clientChan, err := rpc.Dial("tcp", client.all_ip[client.Coordinator])
	if err != nil {
		fmt.Printf("client connection with server %v error\n", client.Coordinator)
	}
	client.rpcChan = clientChan

	// readfilename := []byte("read.txt")
	writefilename := []byte("master")
	writecontents := []byte(id_arg)

	time.Sleep(time.Second * 5)
	//client read request
	// client.SendReadRequest(readfilename)

	//client write request
	go client.ConsistentRead(writefilename)
	client.Write(writefilename, writecontents)

	time.Sleep(time.Second * 5)

	client.SendKeepAlive(client.Coordinator)

}
