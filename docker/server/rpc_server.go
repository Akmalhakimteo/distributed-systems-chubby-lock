package main

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

type Listener int

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

type MessageFileTransfer struct {
	SenderID int
	Msg      string
	Bytedata []byte
}

type Client struct {
	ID               int
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
	ID     int
	// all_ip [3]string
	all_ip      [5]string
	Coordinator int
	electing    bool
	// rpcChan     [3]*rpc.Client //connection channels within servers
	rpcChan     [5]*rpc.Client
	Coord_chng  bool
	dbfilename  string
	written     bool
	writing     bool
	block       bool
	initialized bool
	lock        Maplock
}

type Maplock struct {
	mu    sync.Mutex
	locks map[string]lockCtr //string is the file name, lockctr is the lock object
}

type lockCtr struct {
	clientID   int //-1 when no client holding
	sequenceNo int
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
		// msg := "Received Read Successful"
		data := GetValueFromDB(node.dbfilename, request.Filename)
		// should wait for propagation
		*reply = Reply{data}
	} else if request.Write == 1 {
		start_time := time.Now()
		fmt.Printf("Received Write Request from Client: %v  for file %v with contents: %v\n", request.SenderID, string(request.Filename), string(request.Filecontent))
		var msg string
		// Checks if request for propagation is up
		if node.block {
			msg = "Received Write Failed"
			*reply = Reply{msg}
			return nil
		}
		success := node.clientWriteReq(ClientRequest{node.Coordinator, 1, request.Filename, request.Filecontent})
		// wait for propagation
		for {
			if success {
				go IterateValuesDB(node.dbfilename)
				msg = "Received Write Successful"
				//Uncomment to test multiple clients acquiring the same lock
				// time.Sleep(10 * time.Second)
				ReleaseLock(request, msg)
				break
			}
			//TODO: implement fail msg
			if success == false {
				msg = "Received Write Failed"
				log.Println("Write failed")
				go node.RunPropogateMaster()
				break
			}
			//Timeout of 10s on client sIDe
			t := time.Now()
			if t.Sub(start_time) > (10*time.Second) || !success {
				msg = "Received Write Failed"
				log.Println("Write failed")

				go node.RunPropogateMaster()
				//TODO: RELEASE LOCK
				//Uncomment to test multiple clients acquiring the same lock
				time.Sleep(10 * time.Second)
				ReleaseLock(request, msg)
				break
			}
		}
		*reply = Reply{msg}
	}
	return nil
}

func (l *Listener) Keepalive(c *Client, reply *Reply) error {
	fmt.Printf("Received: KeepAlive Request from client %v\n\n", c.ID)
	concat := "Received KeepAlive reply from server " + strconv.Itoa(c.Coordinator)
	*reply = Reply{concat}
	return nil
}

func (l *Listener) GetCoordinator(ID int, reply *CoordReply) error {
	fmt.Printf("Received: Client %v is asking for new coordinator.\n", ID)
	if node.electing {
		*reply = CoordReply{node.Coordinator, "wait"}
		return nil
	}
	*reply = CoordReply{node.Coordinator, "done"}
	return nil
}

func (l *Listener) Election(msg Message, reply *Message) error {
	fmt.Printf("Received: %v from server %v\n\n", msg.Msg, msg.SenderID)
	if msg.SenderID < node.ID {
		*reply = Message{node.ID, "No"}
		if !node.electing && node.initialized {
			go node.Elect()
		}
	} else if msg.Msg == "I am Coordinator" {
		if msg.SenderID > node.Coordinator {
			log.Println(msg.SenderID, "is the new Coordinator")
			node.Coordinator = msg.SenderID
			node.Coord_chng = true
			*reply = Message{node.ID, "Ack"}
			// go node.ping()
		}
	}
	return nil
}

func (l *Listener) MasterPropogateDB(msg MessageFileTransfer, reply *MessageFileTransfer) error {
	fmt.Printf("Received: %v Master update from server %v\n\n", msg.Msg, msg.SenderID)
	// log.Println("Receiving over bytes: ",msg.Bytedata)
	err := ioutil.WriteFile("Master-db", msg.Bytedata, 0644)
	if err != nil {
		log.Fatal(err)
	}
	CopyMasterFile("Master-db", node.ID)
	log.Println("Updated Database of server", node.ID)
	go IterateValuesDB(node.dbfilename)
	// log.Println("Updated server copy with master copy..Testing get value from  Node",node.ID,node.dbfilename) //uncomment to check for file transfer ok
	// key := [] byte("key roomba")
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
	// if successful in the write
	node.written = true
	//TODO: should wait for DB to finish updating
	node.sample_write(file.FileNames[0], file.FileContents[0])
	//call function that takes input file and updates database
	*reply = Reply{"ACK"}
	return nil
}

func makeNode(ID int) *Node {
	// all_ip := [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}
	all_ip := [5]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234", "172.22.0.5:1234", "172.22.0.6:1234"}
	Coordinator := -1
	// Coordinator := 4
	electing := false
	// var rpcChan [3]*rpc.Client
	var rpcChan [5]*rpc.Client
	dbfilename := InitializeDB(ID)
	written := false
	writing := false
	block := false
	initialized := false
	c := Maplock{locks: make(map[string]lockCtr)}
	curr_node := Node{ID, all_ip, Coordinator, electing, rpcChan, false, dbfilename, written, writing, block, initialized, c}

	go curr_node.connect_all()

	return &curr_node
}

func (n *Node) connect_all() {
	for ind, curr_ip := range n.all_ip {
		// Skip sending msg to self
		if ind == n.ID {
			continue
		}
		go n.connect(ind, curr_ip)
	}

	// pause to allow for connection to establish
	time.Sleep(time.Second)

	coord, ind := n.GetCoordinator()
	log.Println(coord)
	if coord == -1 {
		log.Println("all nodes are just initialized")
	} else {
		// In the case where master server dies and restarts before a new master is elected
		if coord == n.ID {
			// pause to allow other nodes to re-establish connection
			time.Sleep(5 * time.Second)
			// get db from any node
			n.getMasterProp(ind)
			// propagate db to all other node
			n.RunPropogateMaster()
		} else {
			n.getMasterProp(coord)
		}
	}
	if !n.electing {
		go n.Elect()
	}
	n.initialized = true
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

func MasterSendPropogate(msg MessageFileTransfer, rpcChan *rpc.Client) { //Master Send DB to SINGLE server
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
	// if n.Coordinator == n.ID {
	// 	return
	// }
	n.electing = true

	// Establish connection with all higher ip servers
	fmt.Println("Server:", n.ID, "is starting Election")
	// Check with other higher ID servers
	if n.ID+1 < len(n.all_ip) {
		for _, curr_connect := range n.rpcChan[n.ID+1:] {
			if curr_connect != nil {
				log.Println("Server", n.ID, "is requesting to be Coordinator")
				// Start connection protocol
				reply := send_elect(Message{n.ID, "Can I be Coordinator?"}, curr_connect)
				log.Println("Server", n.ID, "Received reply:", reply)
				if reply == "No" {
					n.electing = false
					<-time.After(time.Second * 2)
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
	// If you are the highest ID server/higher ID server dID not reply, broadcast that you are the master
	for ind, curr_connect := range n.rpcChan {
		// Skip sending msg to self
		if ind == n.ID {
			continue
		}
		if curr_connect == nil {
			continue
		}
		// Uncomment this section to test the scenario where node 2 (master) dies before sending "I am Coordinator" message
		// Need to manually stop the process/docker container after election
		// if n.ID == 2 {
		// 	return
		// }
		go send_elect(Message{n.ID, "I am Coordinator"}, curr_connect)
	}
	// Set self to be coordinator
	n.Coordinator = n.ID
	n.RunPropogateMaster() //uncomment to simulate
	// Stop election

	n.electing = false
}

func (n *Node) RunPropogateMaster() { //Yx Call this only when you want to propogate Master DB to ALL servers
	// key := [] byte("key roomba") //uncomment to check for file transfer ok
	// value := [] byte("value samba")
	// WriteToDB(n.dbfilename,key,value)
	masterFile, err := os.Open(n.dbfilename)
	if err != nil {
		log.Fatal(err)
	}
	masterFileInBytes, err := ioutil.ReadAll(masterFile)
	if err != nil {
		log.Fatal(err)
	}
	sendThis := MessageFileTransfer{n.ID, "Servers, follow my master copy", masterFileInBytes}
	for _, curr_connect := range n.rpcChan {
		if curr_connect != nil {
			MasterSendPropogate(sendThis, curr_connect)
		}
	}
}

func (n *Node) ping(ind int) {
	// if n.Coordinator == n.ID {
	// 	return
	// }
	log.Println("Server", n.ID, "is pinging Server:", ind)
	// Start connection protocol
	for {
		<-time.After(5 * time.Second)
		// ind := n.Coordinator
		reply := send(Message{n.ID, "ping"}, n.rpcChan[ind])
		if reply == "error" {
			// reset connection to nil
			n.rpcChan[ind] = nil
			// attempt to re-establish connection within 5s
			go n.connect(ind, n.all_ip[ind])
			// Start election if not already in election
			if !node.electing && ind == n.Coordinator && n.initialized {
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
	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 5 * time.Second}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("DB Initialized for server", nodenumber_str)
	bucketname_byte := []byte("bucket")
	err = db.Update(func(tx *bolt.Tx) error { //Init Bucket
		_, err := tx.CreateBucketIfNotExists(bucketname_byte)
		if err != nil {
			return err
		}
		fmt.Println("Bucket Initialized for server", nodenumber_str)
		return nil
	})
	return dbfilename
}

func WriteToDB(dbfilename string, key, value []byte) error { //if Key-value alerady exists, the value will get updated

	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 5 * time.Second}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
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

func GetValueFromDB(dbfilename string, key []byte) string {
	var new_val string
	bucketname_byte := []byte("bucket")
	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
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
			new_val = string(val)
		} else {
			fmt.Println("Key value does not exist")
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return new_val
}

func IterateValuesDB(dbfilename string) {
	fmt.Println("DB contents of", dbfilename)
	bucketname_byte := []byte("bucket")
	db, err := bolt.Open(dbfilename, 0666, &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true}) //Bolt obtains file lock on data file so multiple processes cannot open same database at the same time. timeout prevents indefinite wait
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(bucketname_byte)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
}

func CopyMasterFile(masterDBfilename string, currentServerNodenumber int) {
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
		if ind == n.ID {
			continue
		}
		if curr_connect == nil {
			continue
		}
		reply := sendWriteData(DatabaseData{n.ID, [][]byte{d.Filename}, [][]byte{d.Filecontent}}, curr_connect)
		log.Println("Server", d.SenderID, "Received reply:", reply)
		if reply == "FAIL" {
			return false
		} else if reply == "ACK" {
			count++
		}
	}
	// if count == 2 {
	// 	// if count ==4{  //UNCOMMENT IF SERVERS == 5
	// 	n.sample_write(d.Filename, d.Filecontent)
	// 	return true
	// }
	if count == 4 {
		n.sample_write(d.Filename, d.Filecontent)
		return true
	}
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
	if n.Coordinator != n.ID {
		return
	}
	//master fetch own db

	//call channels with other servers
	//send through db and wait for ack

	//only continue on when all ack

}

// Lock locks a mutex with the given name. If it doesn't exist, one is created
func (l *Listener) TryAcquire(request ClientRequest, reply *Reply) error {
	fmt.Println("Client is trying to acquire the lock for file", string(request.Filename))
	// log.Println(node.lock)
	node.lock.mu.Lock() //lock the map which works now
	// log.Println(node.lock)
	lock, exist := node.lock.locks[string(request.Filename)]
	node.lock.locks[string(request.Filename)] = lock
	if !exist { //create lock
		lock.clientID = request.SenderID
		*reply = Reply{"You can have the lock"}
		fmt.Println("Client ", request.SenderID, " has been granted the lock")
	} else {
		if lock.clientID == -1 {
			//no one has the lock
			lock.clientID = request.SenderID
			*reply = Reply{"You can have the lock"}
			fmt.Println("Client ", request.SenderID, " has been granted the lock")
		} else {
			*reply = Reply{"Someone else has the lock"}
			fmt.Println("Client ", request.SenderID, " has been denied the lock")
		}
	}
	node.lock.mu.Unlock() //unlock to make sure next write request goes thru
	// else {

	// 	lock.clientID = request.SenderID
	// 	lock.sequenceNo = 0
	// 	*reply = Reply{"You can have the lock"}
	// 	fmt.Println("Client ", request.SenderID, " has been granted the lock")
	// }
	return nil
}

// Unlock unlocks the mutex with the given name
func ReleaseLock(request ClientRequest, msg string) error {
	if msg == "Received Write Successful" {
		lock := node.lock.locks[string(request.Filename)]
		lock.clientID = -1
		lock.sequenceNo++
		fmt.Println("Lock has been released after successful write. New sequence number for file ", string(request.Filename), " is ", lock.sequenceNo)
	} else {
		lock := node.lock.locks[string(request.Filename)]
		lock.clientID = -1 //release lock still but dont update sequenceNo bc it's not a successful write
		fmt.Println("Lock has been released due to unsuccessful write, sequence number for file ", string(request.Filename), " is ", lock.sequenceNo)
	}
	return nil
}

var node *Node

func main() {
	// Uncomment this to test how a higher ID node becomes master server after restarting
	// time.Sleep(5*time.Second)

	// addy, err := net.ResolveTCPAddr("tcp", "172.20.0.2:12345")
	// if err != nil {
	//   log.Fatal(err)
	// }

	// Get ip addr
	// Uncomment to pause to allow for new master to be elected
	// time.Sleep(5*time.Second)
	ip_addr := os.Args[1]
	ID_arg := os.Args[2]
	ID, _ := strconv.Atoi(ID_arg)

	log.Println("Server", ID, "is running")

	node = makeNode(ID)

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

func (n *Node) getMasterProp(coord int) {
	log.Println("getting master db")
	// ask if master has updated
	var reply Message
	// ensure that channel is not nil
	curr_chan := n.rpcChan[coord]
	if curr_chan != nil {
		curr_chan.Call("Listener.CheckMasterDB", Message{n.ID, ""}, &reply)
		if reply.Msg == "not written" {
			log.Println("master db not written so no need to get master db")
			return
		} else if reply.Msg == "wait" {
			time.Sleep(5 * time.Second)
			n.getMasterProp(coord)
			return
		} else if reply.Msg == "sent" {
			log.Println("master says he sent the db already")
		}
	} else {
		log.Println("no connection to:", coord)
	}
}

// get the new coordinator
func (n *Node) GetCoordinator() (int, int) {
	var CoordinatorReply CoordReply
	newCoord := -1
	index := 0
	for ind, curr_connect := range n.rpcChan {
		if ind == n.ID || curr_connect == nil {
			continue
		}
		curr_connect.Call("Listener.GetCoordinator", n.ID, &CoordinatorReply)
		log.Printf(CoordinatorReply.Data)
		if CoordinatorReply.Data == "wait" {
			time.Sleep(5 * time.Second)
			continue
		}
		if CoordinatorReply.Data == "" {
			continue
		}
		newCoordinatorInt := CoordinatorReply.Coord
		newCoord = newCoordinatorInt
		index = ind
		if newCoord == -1 {
			continue
		} else {
			return newCoord, index
		}
	}
	return newCoord, index
}

func (l *Listener) CheckMasterDB(msg Message, reply *Message) error {
	log.Println("check if master has a written DB")
	if !node.written {
		// not written thus no need to send updated DB
		*reply = Message{node.ID, "not written"}
		return nil
	}

	// block all future writes using semaphore
	node.block = true

	if node.electing || node.writing {
		// written DB but in the mIDdle of election or writting
		// tell server to wait
		*reply = Message{node.ID, "wait"}
		return nil
	} else {
		// Send the DB over to the server
		masterFile, err := os.Open(node.dbfilename)
		if err != nil {
			log.Println(err)
		}
		masterFileInBytes, err := ioutil.ReadAll(masterFile)
		if err != nil {
			log.Println(err)
		}
		sendThis := MessageFileTransfer{node.ID, "Servers, follow my master copy", masterFileInBytes}
		MasterSendPropogate(sendThis, node.rpcChan[msg.SenderID])
		*reply = Message{node.ID, "sent"}
		node.block = false
	}
	return nil
}
