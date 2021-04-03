package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
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

func SendKeepAlive(client *Client, serverInt int) {
	var KeepAliveReply Reply
	keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
	time.Sleep(time.Second * 5)
	log.Printf(KeepAliveReply.Data)
	if keepalive_err != nil {
		var newCoordinatorInt int
		fmt.Printf("server node %v died\n", serverInt)
		// get new coordinator
		for ind, curr_ip := range client.all_ip {
			clientChan, err := rpc.Dial("tcp", curr_ip)
			if err != nil {
				fmt.Printf("client connection with server %v error\n", ind)
				continue
			}
			client.rpcChan = clientChan

			var CoordinatorReply Reply
			client.rpcChan.Call("Listener.GetCoordinator", &client, &CoordinatorReply)
			time.Sleep(time.Second * 5)
			log.Printf(CoordinatorReply.Data)
			newCoordinator := CoordinatorReply.Data[len(CoordinatorReply.Data)-1:]
			newCoordinatorInt, err = strconv.Atoi(newCoordinator)
			client.Coordinator = newCoordinatorInt
			if err != nil {
				fmt.Println("hello?")
			}
			clientChan, err = rpc.Dial("tcp", client.all_ip[newCoordinatorInt])
			if err != nil {
				fmt.Printf("client connection with server %v error\n", client.Coordinator)
			}
			client.rpcChan = clientChan
			SendKeepAlive(client, newCoordinatorInt)
			break
		}
	} else {
		SendKeepAlive(client, serverInt)
	}
}

func main() {

	//TODO: Client needs to communicate with chubby cell to find out coordinator

	log.Printf("Client is created")

	client := Client{id: 0, Coordinator: 2, all_ip: [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}}
	clientChan, err := rpc.Dial("tcp", client.all_ip[client.Coordinator])
	if err != nil {
		fmt.Printf("client connection with server %v error\n", client.Coordinator)
	}
	client.rpcChan = clientChan
	SendKeepAlive(&client, client.Coordinator)
	// clientChan, err := rpc.Dial("tcp", client.all_ip[2])
	// if err != nil {
	// 	log.Fatal("client connection with server 2 error")
	// }
	// client.rpcChan = clientChan

	// //Client first exchange keepalives with Node/Server 2
	// for {
	// 	//keepalive
	// 	var KeepAliveReply Reply
	// 	keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
	// 	time.Sleep(time.Second * 5)
	// 	log.Printf(KeepAliveReply.Data)

	// 	if keepalive_err != nil {
	// 		fmt.Println("server node died")

	// 		// Client needs to communicate with chubby cell to find out new coordinator
	// 		for ind, curr_ip := range client.all_ip {
	// 			client.rpcChan, err = rpc.Dial("tcp", curr_ip)
	// 			if err != nil {
	// 				log.Println("Cannot connect with Server: ", ind)
	// 				continue
	// 			}

	// 			var CoordinatorReply Reply
	// 			client.rpcChan.Call("Listener.GetCoordinator", &client, &CoordinatorReply)
	// 			time.Sleep(time.Second * 5)
	// 			log.Printf(CoordinatorReply.Data)
	// 			newCoordinator := CoordinatorReply.Data[len(CoordinatorReply.Data)-1:]
	// 			newCoordinatorInt, err := strconv.Atoi(newCoordinator)
	// 			if err != nil {
	// 				fmt.Println("hello?")
	// 			}
	// 			// Establish new connection with new coordinator
	// 			for {
	// 				var KeepAliveReply Reply
	// 				clientChan, err := rpc.Dial("tcp", client.all_ip[newCoordinatorInt])
	// 				if err != nil {
	// 					log.Fatal("client connection with server 1 error")
	// 				}
	// 				client.rpcChan = clientChan
	// 				client.Coordinator = newCoordinatorInt
	// 				keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
	// 				if keepalive_err != nil {
	// 					log.Fatal(keepalive_err)
	// 					log.Fatal("server node died again")
	// 				}
	// 				time.Sleep(time.Second * 5)
	// 				log.Printf(KeepAliveReply.Data)
	// 			}
	// 		}
	// 	}
	// }
	// i := 2
	// for {
	// 	clientChan, err := rpc.Dial("tcp", client.all_ip[i])
	// 	if err != nil {
	// 		log.Println("Cannot connect with Server: ", i)
	// 	}
	// 	client.rpcChan = clientChan

	// 	//keepalive
	// 	var KeepAliveReply Reply
	// 	keepalive_err := client.rpcChan.Call("Listener.Keepalive", &client, &KeepAliveReply)
	// 	time.Sleep(time.Second * 5)
	// 	log.Printf(KeepAliveReply.Data)

	// 	if keepalive_err != nil {
	// 		fmt.Println("server node died")
	// 		var CoordinatorReply Reply
	// 		client.rpcChan.Call("Listener.GetCoordinator", &client, &CoordinatorReply)
	// 		log.Printf(CoordinatorReply.Data)
	// 		newCoordinator := CoordinatorReply.Data[len(CoordinatorReply.Data)-1:]
	// 		newCoordinatorInt, err := strconv.Atoi(newCoordinator)
	// 		if err != nil {
	// 			fmt.Printf("Error getting new coordinator %v", newCoordinatorInt)
	// 		}
	// 		i = newCoordinatorInt
	// 		continue
	// 	}
	// }

}
