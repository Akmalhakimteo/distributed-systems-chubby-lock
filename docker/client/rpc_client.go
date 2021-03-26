package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

type Reply struct {
	Data string
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

func main() {

	//TODO: Client needs to communicate with chubby cell to find out coordinator

	log.Printf("Client is created")

	client := Client{id: 0, Coordinator: 2, all_ip: [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}}
	clientChan, err := rpc.Dial("tcp", client.all_ip[2])
	if err != nil {
		log.Fatal("client connection with server 2 error")
	}
	client.rpcChan = clientChan

	// Client first exchange keepalives with Node/Server 2
	for {
		//keepalive
		var KeepAliveReply Reply
		keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
		time.Sleep(time.Second * 5)
		log.Printf(KeepAliveReply.Data)
		if keepalive_err != nil {
			// TODO: Client needs to communicate with chubby cell to find out coordinator
			fmt.Println("server node died")
			break
		}
	}

	// TODO: Client needs to first communicate with chubby cell to find out new coordinator
	// Client then exchange keepalives with Node/Server 1
	for {
		var KeepAliveReply Reply
		clientChan, err := rpc.Dial("tcp", client.all_ip[1])
		if err != nil {
			log.Fatal("client connection with server 1 error")
		}
		client.rpcChan = clientChan
		client.Coordinator = 1
		keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
		if keepalive_err != nil {
			log.Fatal(keepalive_err)
			log.Fatal("server node died again")
		}
		time.Sleep(time.Second * 5)
		log.Printf(KeepAliveReply.Data)
	}
}
