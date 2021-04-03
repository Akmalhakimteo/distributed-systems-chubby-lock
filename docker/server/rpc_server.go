package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
	bolt "go.etcd.io/bbolt"
	"io"
)

type Listener int

type Reply struct {
	Data string
}

type Message struct {
	SenderID int
	Msg      string
}

type Client struct {
	id               int
	Coordinator      int
	rpcChan          *rpc.Client
	keepAliveRequest bool
}

type Node struct {
	id     int
	all_ip [3]string
	// all_ip      [5]string
	Coordinator int
	electing    bool
	rpcChan     [3]*rpc.Client //connection channels within servers
	// rpcChan     [5]*rpc.Client
	Coord_chng bool
	// revChan chan Message
	// replyChan chan Message
	// ClientChan chan c.SelfIntroduction
	// peers []Node
	// quitElect chan int
	// killNode chan int
	// AllClients []chan c.File
	// FileChan chan c.File
	// database map[string]Lock
}

func (l *Listener) GetLine(msg Message, reply *Reply) error {
	// rv := string(line)
	fmt.Printf("Received: %v from server %v\n\n", msg.Msg, msg.SenderID)
	concat := "received: " + msg.Msg
	*reply = Reply{concat}
	return nil
}

func (l *Listener) Keepalive(c *Client, reply *Reply) error {
	fmt.Printf("Received: KeepAlive Request from client %v\n\n", c.id)
	concat := "Received KeepAlive reply from server " + strconv.Itoa(c.Coordinator)
	*reply = Reply{concat}
	return nil
}

func (l *Listener) GetCoordinator(c *Client, reply *Reply) error {
	fmt.Printf("Received: Client %v is asking for new coordinator.\n", c.id)
	concat := "New Coordinator is: " + strconv.Itoa(node.Coordinator)
	*reply = Reply{concat}
	return nil
}

func (l *Listener) Election(msg Message, reply *Message) error {
	fmt.Printf("Received: %v from server %v\n\n", msg.Msg, msg.SenderID)
	if msg.SenderID < node.id {
		*reply = Message{node.id, "No"}
		if !node.electing {
			go node.Elect()
		}
	} else if msg.Msg == "I am Coordinator" {
		log.Println(msg.SenderID, "is the new Coordinator")
		node.Coordinator = msg.SenderID
		node.Coord_chng = true
		*reply = Message{node.id, "Ack"}
		go node.ping()
	}
	return nil
}

func makeNode(id int) *Node {
	all_ip := [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}
	// all_ip := [5]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234", "172.22.0.5:1234", "172.22.0.6:1234"}
	Coordinator := 2
	// Coordinator := 4
	electing := false
	var rpcChan [3]*rpc.Client
	// var rpcChan [5]*rpc.Client
	curr_node := Node{id, all_ip, Coordinator, electing, rpcChan, false}

	go curr_node.connect_all()

	return &curr_node
}

func (n *Node) connect_all() {
	for ind, curr_ip := range n.all_ip {
		// Skip sending msg to self
		if ind == n.id {
			continue
		}
		curr_connect, err := rpc.Dial("tcp", curr_ip)
		if err != nil {
			log.Println("Cannot connect with server:", ind)
			// curr_connect.Close()
			continue
		}
		n.rpcChan[ind] = curr_connect
	}
	log.Println("rpcChan:", n.rpcChan)
	go n.ping()
}

func send(msg Message, rpcChan *rpc.Client, send_id int) string {
	var reply Reply
	err := rpcChan.Call("Listener.GetLine", msg, &reply)
	if err != nil {
		if err.Error() == "connection is shut down" {
			log.Println("connection is shut down, starting election")
			if !node.electing {
				node.Coord_chng = false
				go node.Elect()
			}
			return "error"
		}
		log.Printf("type: %T\n", err)
		// log.Fatal(err)
	}
	log.Printf("Reply: %v, Data: %v\n", reply, reply.Data)
	return reply.Data
}

func send_elect(msg Message, rpcChan *rpc.Client) string {
	var reply Message
	start_time := time.Now()
	for {
		err := rpcChan.Call("Listener.Election", msg, &reply)
		if err != nil {
			if err.Error() == "connection is shut down" {
				// log.Println("connection is shut down")
				t := time.Now()
				if t.Sub(start_time) > (5 * time.Second) {
					return ""
				}
				continue
			}
			log.Printf("type: %T\n", err)
			return ""
			// log.Fatal(err)
		}
		break
	}
	log.Printf("Reply: %v, Message: %v\n", reply, reply.Msg)
	return reply.Msg
}

func (n *Node) Elect() {
	if n.Coordinator == n.id {
		return
	}
	n.electing = true

	// Establish connection with all higher ip servers
	fmt.Println("Server:", n.id, "is starting Election")
	// Check with other higher id servers
	if n.id+1 < len(n.all_ip) {
		for _, curr_connect := range n.rpcChan[n.id+1:] {
			if curr_connect != nil {
				log.Println("Server", n.id, "is requesting to be Coordinator")
				// Start connection protocol
				reply := send_elect(Message{n.id, "Can I be Coordinator?"}, curr_connect)
				log.Println("Server", n.id, "Received reply:", reply)
				if reply == "No" {
					n.electing = false
					<-time.After(time.Second * 5)
					log.Println("Has Coordinator changed?", n.Coord_chng)
					if !n.Coord_chng && !n.electing {
						go n.Elect()
						return
					}
					return
				}
			}
		}
	}
	// If you are the highest id server/higher id server did not reply, broadcast that you are the master
	for ind, curr_connect := range n.rpcChan {
		// Skip sending msg to self
		if ind == n.id {
			continue
		}
		if curr_connect == nil {
			continue
		}
		send_elect(Message{n.id, "I am Coordinator"}, curr_connect)
	}
	// Set self to be coordinator
	n.Coordinator = n.id
	// Stop election
	n.electing = false
}

func (n *Node) ping() {
	if n.Coordinator == n.id {
		return
	}
	log.Println("Server", n.id, "is pinging Coordinator:", n.Coordinator)
	// Start connection protocol
	for {
		<-time.After(5 * time.Second)
		reply := send(Message{n.id, "ping"}, n.rpcChan[n.Coordinator], n.Coordinator)
		if reply == "error" {
			return
		}
		log.Printf("Received reply from Coordinator: %v", reply)
	}
}

func InitializeDB(nodenumber int) string {
	nodenumber_str := strconv.Itoa(nodenumber) //Init DB
	var dbname_temp = "Node-db"
	dbfilename := dbname_temp[:4] + nodenumber_str + dbname_temp[4:]
	db, err:= bolt.Open(dbfilename,0666,&bolt.Options{Timeout: 1 * time.Second})  //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err!= nil{
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("DB Initialized for server", nodenumber_str)
	return dbfilename
}

func WriteToDB(dbfilename string, key,value [] byte){   //if Key-value alerady exists, the value will get updated

	db, err:= bolt.Open(dbfilename,0666,&bolt.Options{Timeout: 1 * time.Second})  //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err!= nil{
		log.Fatal(err)
	}
	defer db.Close()

	bucketname_byte := []byte("bucket")
	err = db.Update(func(tx *bolt.Tx) error {  //Init Bucket
        bucket, err := tx.CreateBucketIfNotExists(bucketname_byte)
        if err != nil {
            return err
        }
		err = bucket.Put(key, value)
        if err != nil {
            return err
        }
		fmt.Println("Key value successfully written to",dbfilename)
		return nil
    })

	if err != nil {
        log.Fatal(err)
    }
}

func GetValueFromDB(dbfilename string, key [] byte){
	bucketname_byte := []byte("bucket")
	db, err:= bolt.Open(dbfilename,0666,&bolt.Options{Timeout: 1 * time.Second})  //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err!= nil{
		log.Fatal(err)
	}
	defer db.Close()
	// // retrieve the data
	err = db.View(func(tx *bolt.Tx) error {
        bucket := tx.Bucket(bucketname_byte)
        if bucket == nil {
            return fmt.Errorf("Bucket not found")
        }
        val := bucket.Get(key)
		if val!=nil{
			fmt.Println(string(val))
		}else{
			fmt.Println("Key value does not exist")
		}
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }
}

func PropogateMaster(masterDBfilename string,currentServerNodenumber int){
	nodenumber_str := strconv.Itoa(currentServerNodenumber) 
	var dbname_temp = "Node-db"
	dbfilename := dbname_temp[:4] + nodenumber_str + dbname_temp[4:]
	os.Remove(dbfilename)
	
	sourceFile, err := os.Open(masterDBfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceFile.Close()
	
	// Create new file
	newFile, err := os.Create(dbfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	bytesCopied, err := io.Copy(newFile, sourceFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Master copy propogated. Copied %d bytes.", bytesCopied)

}

var node *Node

func main() {
	// addy, err := net.ResolveTCPAddr("tcp", "172.20.0.2:12345")
	// if err != nil {
	//   log.Fatal(err)
	// }

	// Get ip addr
	ip_addr := os.Args[1]
	id_arg := os.Args[2]
	id, _ := strconv.Atoi(id_arg)

	log.Println("Server", id, "is running")

	node = makeNode(id)
	dbfilename := InitializeDB(id)
	key := []byte("Marco")
    value := []byte("Polo")
	WriteToDB(dbfilename,key,value)
	GetValueFromDB(dbfilename,key)

	// inbound, err := net.Listen("tcp", "172.22.0.3:1234")
	inbound, err := net.Listen("tcp", ip_addr)
	if err != nil {
		log.Fatal(err)
	}
	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)

}
