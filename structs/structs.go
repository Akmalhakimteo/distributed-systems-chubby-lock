package structs

type Client struct{
	id int
	RevChan chan Message
	// recvKeepAlive chan KeepAlive
	masterServer int
	AllServer []Node

	// revChan chan Message
	// replyChan chan Message
	// peers []Node
	// quitElect chan int
	// killNode chan int
	// coordinator int
}

type Node struct{
	Id int
	RevChan chan Message
	ReplyChan chan Message
	ClientChan chan SelfIntroduction
	Peers []Node
	QuitElect chan int
	KillNode chan int
	Coordinator int
	AllClient []Client
}

type Message struct{
	SenderID int
	Msg string
}

type SelfIntroduction struct{
	SenderID int
	RevChan chan Message
	Msg string
}