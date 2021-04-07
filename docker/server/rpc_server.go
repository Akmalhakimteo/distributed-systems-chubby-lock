package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
	bolt "go.etcd.io/bbolt"
	"io/ioutil"

)

type Listener int

type Reply struct {
	Data string
}

type Message struct {
	SenderID int
	Msg      string
}

type MessageFileTransfer struct {
	SenderID int
	Msg      string
	Bytedata [] byte
}

type Client struct {
	id               int
	Coordinator      int
	rpcChan          *rpc.Client
	keepAliveRequest bool
}

type ClientRequest struct {
	SenderID    int
	Write       int // 1 = write request, 0 = read request
	Filename    []byte
	Filecontent []byte //empty for read request
}

type DatabaseData struct {
	SenderID     int
	FileNames    [][]byte
	FileContents [][]byte
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
	dbfilename string
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

func (l *Listener) GetRequest(request ClientRequest, reply *Reply) error {
	if request.Write == 0 {
		fmt.Printf("Received Read request from Client: %v for file: %v\n", request.SenderID, string(request.Filename))
		msg := "Received Read Successful"
		// should wait for propagation
		*reply = Reply{msg}
	} else if request.Write == 1 {
		start_time := time.Now()
		fmt.Printf("Received Write Request from Client: %v  for file %v with contents: %v\n", request.SenderID, string(request.Filename), string(request.Filecontent))
<<<<<<< HEAD
=======
		// m := New() //create a new lock
		// m.Lock(string(request.Filename))
		// fmt.Printf(v)
		// fmt.Printf(string(request.Filename), " is locked")

>>>>>>> e7a029a8f831f4b5ad5ae7eee0f92698cd52f9cf
		var msg string
		success := node.clientWriteReq(ClientRequest{node.Coordinator, 1, request.Filename, request.Filecontent})
		// wait for propagation
		for {
			if success {
				msg = "Received Write Successful"
<<<<<<< HEAD
=======
				//TODO: RELEASE LOCK
				ReleaseLock(request, msg)
>>>>>>> e7a029a8f831f4b5ad5ae7eee0f92698cd52f9cf
				break
			}
			//TODO: implement fail msg
			//if success == false {}
			//Timeout of 10s on client side
			t := time.Now()
			if t.Sub(start_time) > (10*time.Second) || !success {
				msg = "Received Write Failed"
<<<<<<< HEAD
=======
				//TODO: RELEASE LOCK
				ReleaseLock(request, msg)
>>>>>>> e7a029a8f831f4b5ad5ae7eee0f92698cd52f9cf
				break
			}
		}
		*reply = Reply{msg}
	}
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
		if msg.SenderID > node.Coordinator {
			log.Println(msg.SenderID, "is the new Coordinator")
			node.Coordinator = msg.SenderID
			node.Coord_chng = true
			*reply = Message{node.id, "Ack"}
			// go node.ping()
		}
	}
	return nil
}

func (l * Listener) MasterPropogateDB(msg MessageFileTransfer, reply *MessageFileTransfer) error {
	fmt.Printf("Received: %v Master update from server %v\n\n", msg.Msg, msg.SenderID)
	// log.Println("Receiving over bytes: ",msg.Bytedata)
	err := ioutil.WriteFile("Master-db", msg.Bytedata, 0644)
	if err != nil {
		log.Fatal(err)
	}
	CopyMasterFile("Master-db",node.id)
	log.Println("Updated Database of server",node.id)
	// log.Println("Updated server copy with master copy..Testing get value from  Node",node.id,node.dbfilename)
	// key := [] byte("key roomba")  //uncomment to check for file transfer ok
	// GetValueFromDB(node.dbfilename,key)
	return nil
}





//message will contain new file data
func (l *Listener) MasterWrite(file DatabaseData, reply *Reply) error {
	if len(file.FileNames) > 1 {
		fmt.Printf("Received: DB from Master Server %v.\n", file.SenderID)
	} else {
		fmt.Printf("Received: file %v from Master Server %v.\n", file.FileNames[0], file.SenderID)
	}
	//should wait for DB to finish updating
	*reply = Reply{"ACK"}
	return nil
}

func makeNode(id int) *Node {
	all_ip := [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}
	// all_ip := [5]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234", "172.22.0.5:1234", "172.22.0.6:1234"}
	Coordinator := -1
	// Coordinator := 4
	electing := false
	var rpcChan [3]*rpc.Client
	// var rpcChan [5]*rpc.Client
	dbfilename := InitializeDB(id)
	curr_node := Node{id, all_ip, Coordinator, electing, rpcChan, false, dbfilename}

	go curr_node.connect_all()
	
	

	return &curr_node
}



func (n *Node) connect_all() {
	for ind, curr_ip := range n.all_ip {
		// Skip sending msg to self
		if ind == n.id {
			continue
		}
		go n.connect(ind, curr_ip)
		// curr_connect, err := rpc.Dial("tcp", curr_ip)
		// if err != nil {
		// 	log.Println("Cannot connect with server:", ind)
		// 	// curr_connect.Close()
		// 	continue
		// }
		// n.rpcChan[ind] = curr_connect
	}
	// log.Println("rpcChan:", n.rpcChan)
	// go n.ping()
	go n.Elect()
}

func send(msg Message, rpcChan *rpc.Client) string {
	var reply Reply
	err := rpcChan.Call("Listener.GetLine", msg, &reply)
	if err != nil {
		if err.Error() == "connection is shut down" {
			log.Println("connection is shut down")
			rpcChan.Close()
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

func MasterSendPropogate(msg MessageFileTransfer, rpcChan  *rpc.Client ){ //Master Send DB to SINGLE server
	var reply Message
	for {
		err := rpcChan.Call("Listener.MasterPropogateDB", msg, &reply)
		log.Printf("Master ping to servers: Follow master copy")
		if err != nil {
			log.Printf("Error occured while pinging for propogate")
		}
		break
	}
}





func (n *Node) Elect() {
	// pause to allow for connection to establish
	time.Sleep(5 * time.Second)
	// if n.Coordinator == n.id {
	// 	return
	// }
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
		go send_elect(Message{n.id, "I am Coordinator"}, curr_connect)
	}
	// Set self to be coordinator
	n.Coordinator = n.id
	// n.RunPropogateMaster() //uncomment to simulate
	// Stop election
	n.electing = false
}

func (n *Node) RunPropogateMaster(){ //Yx Call this only when you want to propogate Master DB to ALL servers
	// key := [] byte("key roomba") //uncomment to check for file transfer ok
	// value := [] byte("value samba")
	// WriteToDB(n.dbfilename,key,value)
	masterFile, err := os.Open(n.dbfilename)
	if err != nil {
		log.Fatal(err)
	}
	masterFileInBytes,err := ioutil.ReadAll(masterFile)
	if err != nil {
		log.Fatal(err)
	}
	sendThis := MessageFileTransfer{n.id,"Servers, follow my master copy",masterFileInBytes}
	for _, curr_connect := range n.rpcChan {
		if curr_connect != nil {
			MasterSendPropogate(sendThis,curr_connect)
		}
	}
}

func (n *Node) ping(ind int) {
	// if n.Coordinator == n.id {
	// 	return
	// }
	log.Println("Server", n.id, "is pinging Server:", ind)
	// Start connection protocol
	for {
		<-time.After(5 * time.Second)
		// ind := n.Coordinator
		reply := send(Message{n.id, "ping"}, n.rpcChan[ind])
		if reply == "error" {
			// reset connection to nil
			n.rpcChan[ind] = nil
			// attempt to re-establish connection within 5s
			go n.connect(ind, n.all_ip[ind])
			// Start election if not already in election
			if !node.electing && ind == n.Coordinator {
				log.Println("starting election")
				node.Coordinator = -1
				node.Coord_chng = false
				go node.Elect()
			}
			return
		}
		// log.Printf("Received reply from Server: %v", reply)
	}
}

func InitializeDB(nodenumber int) string {
	nodenumber_str := strconv.Itoa(nodenumber) //Init DB
	var dbname_temp = "Node-db"
	dbfilename := dbname_temp[:4] + nodenumber_str + dbname_temp[4:]
	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 1 * time.Second}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("DB Initialized for server", nodenumber_str)
	bucketname_byte := []byte("bucket")
	err = db.Update(func(tx *bolt.Tx) error {  //Init Bucket
        _, err := tx.CreateBucketIfNotExists(bucketname_byte)
        if err != nil {
            return err
        }
		fmt.Println("Bucket Initialized for server", nodenumber_str)
		return nil
    })
	return dbfilename
}

func WriteToDB(dbfilename string, key,value [] byte) error {   //if Key-value alerady exists, the value will get updated

	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 1 * time.Second}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bucketname_byte := []byte("bucket")
	err = db.Update(func(tx *bolt.Tx) error { //Init Bucket
		bucket, err := tx.CreateBucketIfNotExists(bucketname_byte)
		if err != nil {
			return err
		}
		err = bucket.Put(key, value)
		if err != nil {
			return err
		}
		fmt.Println("Key value successfully written to", dbfilename)
		return nil
	})

	if err != nil {
        log.Fatal(err)
    }
	return err
}

func GetValueFromDB(dbfilename string, key []byte) {
	bucketname_byte := []byte("bucket")
	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 1 * time.Second}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err != nil {
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
		if val != nil {
			fmt.Println(string(val))
		} else {
			fmt.Println("Key value does not exist")
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func CopyMasterFile(masterDBfilename string,currentServerNodenumber int){
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
	os.Remove(masterDBfilename)
    

}

//only master will call this function
//function to handle write request from client
func (n *Node) clientWriteReq(d ClientRequest) bool {
	fmt.Println("Server:", d.SenderID, "is forwarding write to other servers")
	count := 0
	for ind, curr_connect := range n.rpcChan {
		if ind == n.id {
			continue
		}
		if curr_connect == nil {
			continue
		}
		reply := sendWriteData(DatabaseData{n.id, [][]byte{d.Filename}, [][]byte{d.Filecontent}}, curr_connect)
		log.Println("Server", d.SenderID, "Received reply:", reply)
		if reply == "FAIL" {
			return false
		} else if reply == "ACK" {
			count++
		}
	}
	if count == 4 {
		return true
	}
	//something has failed and should call master propogate
	//TODO:master propagate db
	return false
}

func sendWriteData(d DatabaseData, rpcChan *rpc.Client) string {
	var reply Reply
	start_time := time.Now()
	for {
		err := rpcChan.Call("Listener.MasterWrite", d, &reply)
		if err != nil {
			if err.Error() == "connection is shut down" {
				log.Println("server node has died, write will fail")
				return "FAIL"
			}
			t := time.Now()
			//internode time out of 5s
			if t.Sub(start_time) > (5 * time.Second) {
				return "FAIL"
			}
		}
		break
	}
	log.Printf("Reply: %v, Message: %v\n", reply, reply.Data)
	return reply.Data
}

//rollback event
//when master does not receive ACK from all servers
//->they shoud revert to their previous state (copy from master)->master hasnt updated
// called at master and master propogates to all nodes that have sent ACK to previous client write
func (n *Node) masterPropogateDB() {
	if n.Coordinator != n.id {
		return
	}
	//master fetch own db

	//call channels with other servers
	//send through db and wait for ack

	//only continue on when all ack

}

// Lock locks the mutex
// func (l *lockCtr) Lock() {
// 	l.mu.Lock()
// }

// // Unlock unlocks the mutex
// func (l *lockCtr) Unlock() {
// 	l.mu.Unlock()
// }

// // New creates a new Maplock
// func New() *Maplock {
// 	return &Maplock{
// 		locks: make(map[string]*lockCtr),
// 	}
// }

// Lock locks a mutex with the given name. If it doesn't exist, one is created
func (l *Listener) TryAcquire(request ClientRequest, reply *Reply) {
	// // l.mu.Lock()
	// if l.locks == nil {
	// 	l.locks = make(map[string]*lockCtr)
	// }

	// nameLock, exists := l.locks[name]
	// if !exists {
	// 	nameLock = &lockCtr{}
	// 	l.locks[name] = nameLock
	// }
	// // this makes sure that the lock isn't deleted if `Lock` and `Unlock` are called concurrently
	// l.mu.Unlock()

	// // Lock the nameLock outside the main mutex so we don't block other operations
	// nameLock.Lock()
	// Lock := node.lock

	lock, exist := node.lock.locks[string(request.Filename)]
	if exist {
		avail := lock.clientID
		if avail == -1 {
			//no one has the lock
			lock.clientID = request.SenderID
			*reply = Reply{"You can have the lock"}
		} else {
			*reply = Reply{"Someone else has the lock"}
		}
	} else {
		lock.clientID = request.SenderID
		lock.sequenceNo = 0
		*reply = Reply{"You can have the lock"}
	}

}

// Unlock unlocks the mutex with the given name
func ReleaseLock(request ClientRequest, msg string) error {
	if msg == "Received Write Successful" {
		lock := node.lock.locks[string(request.Filename)]
		lock.clientID = -1
		lock.sequenceNo++
	}
	return nil
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

	// inbound, err := net.Listen("tcp", "172.22.0.3:1234")
	inbound, err := net.Listen("tcp", ip_addr)
	if err != nil {
		log.Fatal(err)
	}
	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)


	



}

func (n *Node) sample_write(key []byte, value []byte) {
	WriteToDB(n.dbfilename, key, value)
}

func (n *Node) sample_get(key []byte) {
	GetValueFromDB(n.dbfilename, key)
}

// Attempts to connect with the ip address for 5s, gives up otherwise
func (n *Node) connect(ind int, curr_ip string) {
	start_time := time.Now()
	for {
		curr_connect, err := rpc.Dial("tcp", curr_ip)
		if err != nil {
			end_time := time.Now()
			elapsed := end_time.Sub(start_time)
			if elapsed >= (5 * time.Second) {
				log.Println("Cannot connect with server:", ind, "within 5s")
				return
			}
			continue
		}
		n.rpcChan[ind] = curr_connect
		log.Println("rpcChan:", n.rpcChan)
		go n.ping(ind)
		return
	}
	return
}
