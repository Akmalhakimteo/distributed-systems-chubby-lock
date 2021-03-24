package main
import (
   "log"
   "net/rpc"
)
type Reply struct {
   Data string
}
func main() {
  log.Printf("Client is created")
  client, err := rpc.Dial("tcp", "172.22.0.3:1234")
  if err != nil {
    log.Fatal(err)
  }
  for {
    line := "hello"
    if err != nil {
       log.Fatal(err)
    }
    var reply Reply
    err = client.Call("Listener.GetLine", line, &reply)
    if err != nil {
      log.Fatal(err)
    }
    log.Printf("Reply: %v, Data: %v", reply, reply.Data)
    break
  }
}