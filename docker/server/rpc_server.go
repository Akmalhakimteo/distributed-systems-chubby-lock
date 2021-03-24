package main
import (
   "fmt"
   "log"
   "net"
   "net/rpc"
   "os"
   "strconv"
   "time"
)
type Listener int

type Reply struct {
   Data string
}

type Message struct{
   SenderID int
   Msg string
}

type Node struct{
   id int
   all_ip [3]string
   Coordinator int
   rpcChan *rpc.Client
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
   concat := "received: "+msg.Msg
   *reply = Reply{concat}
   return nil
}

func makeNode(id int) *Node {
   all_ip := [3]string{"172.22.0.7:1234", "172.22.0.3:1234", "172.22.0.4:1234"}
   Coordinator := 2
   curr_node := Node{id, all_ip, Coordinator, nil}

   // allNodes[i].revChan = make(chan Message)
   // allNodes[i].replyChan = make(chan Message)
   // allNodes[i].quitElect = make(chan int)
   // allNodes[i].killNode = make(chan int)
   // allNodes[i].peers = allNodes
   // allNodes[i].Coordinator = numNodes - 1
   // allNodes[i].ClientChan = make(chan c.SelfIntroduction)
   // allNodes[i].database = make(map[string]Lock)

   // for i := 0; i < numNodes; i++ {
   //    go allNodes[i].ping()
   //    go allNodes[i].checkChannel()
   // }

   go curr_node.ping()

   return &curr_node
}

func send(msg Message, rpcChan *rpc.Client, send_id int) string{
   var reply Reply
   err := rpcChan.Call("Listener.GetLine", msg, &reply)
   if err != nil {
      if err.Error() == "connection is shut down"{
         log.Println("yes")
      }
      log.Printf("type: %T\n", err)
      // log.Fatal(err)
   }
   log.Println("Does it reach here?")
   log.Printf("Reply: %v, Data: %v\n", reply, reply.Data)
   return reply.Data
}

func (n *Node) ping() {
   if n.Coordinator == n.id{
      return
   }

   // Establish connection with new Coordinator
   server, err := rpc.Dial("tcp", n.all_ip[n.Coordinator])
   if err != nil {
    log.Fatal(err)
   }
   n.rpcChan = server

   log.Println("Server", n.id, "is pinging Coordinator:", n.Coordinator)
   // Start connection protocol
   for{
      <-time.After(5*time.Second)
      reply := send(Message{n.id, "ping"}, n.rpcChan, n.Coordinator)
      log.Printf("Received reply from Coordinator: %v", reply)
   }
   // Close the connection to server
   n.rpcChan.Close()
   // for {
   //    // fmt.Println("peers: ", n.peers)
   //    fmt.Println("node: ", n.id, "is pinging master server: ", n.Coordinator)
   //    select{
   //    case n.peers[n.Coordinator].revChan<-Message{n.id, "Are you alive?"}:
   //       time.Sleep(time.Second * 3)
   //    case <-time.After(time.Second * 10):
   //       n.elect()
   //    case <-n.killNode:
   //       return
   //    }
   // }
}

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

   go makeNode(id)

   // inbound, err := net.Listen("tcp", "172.22.0.3:1234")
   inbound, err := net.Listen("tcp", ip_addr)
   if err != nil {
     log.Fatal(err)
   }
   listener := new(Listener)
   rpc.Register(listener)
   rpc.Accept(inbound)
}