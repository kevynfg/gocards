package game

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Room struct {

	// clients holds all current clients in this room.
	Clients map[*client]bool

	// join is a channel for clients wishing to join the room.
	Join chan *client

	// leave is a channel for clients wishing to leave the room.
	Leave chan *client

	// forward is a channel that holds incoming messages that should be forwarded to the other clients.
	Forward chan []byte

	// id is the id of the room.
	Id string
}

func GenerateRandomID() string {
	randomIdString := ""
	numbers := make([]int, 20)
	for i := 0; i < 10; i++ {
		randomIdString += strconv.Itoa(numbers[i])
	}
	return randomIdString
}

func NewRoom() *Room {
	return &Room{
		Forward: make(chan []byte),
		Join:    make(chan *client),
		Leave:   make(chan *client),
		Clients: make(map[*client]bool),
		Id:      "",
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Join:
			fmt.Println("client joined room id:", client.room.Id)
			r.Clients[client] = true
		case client := <-r.Leave:
			delete(r.Clients, client)
			close(client.receive)
		case msg := <-r.Forward:
			fmt.Printf("room id: %v, clients: %v\n", r.Id, r.Clients)
			for client := range r.Clients {
				fmt.Printf("sending message to client id: %v on room: %v\n", client.room.Id, r.Id)
				select {
				case client.receive <- msg:
				default:
					delete(r.Clients, client)
					close(client.receive)
				}
			}
		}
		fmt.Printf("room is running with %d clients\n", len(r.Clients))
		fmt.Println("currently running clients: ", r.Clients)
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *Room) ServeHTTP(ctx *gin.Context) {
	if r.Id == "" {
		log.Fatal("room id is required")
		return
	}
	
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	socket, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	
	formatted := fmt.Sprintf("socket address connected %s", socket.RemoteAddr())
	fmt.Println(formatted)
	
	client := &client{
		socket: socket,
		receive: make(chan []byte, messageBufferSize),
		room: r,
	}
	
	r.Join <- client
	
	defer func() { 
		r.Leave <- client
		fmt.Println("client left", client)
	}()
	
	go client.write()
	client.read()
}