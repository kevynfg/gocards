package game

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	qrcode "github.com/skip2/go-qrcode"
)

var url = "http://localhost:8080/ws/room/join/"

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
				fmt.Printf("sending to client: %v on room id: %v with message: %v\n", client.socket.RemoteAddr().String(), client.room.Id, string(msg))
				select {
				case client.receive <- msg:
				default:
					if !r.Clients[client] {
						delete(r.Clients, client)
						close(client.receive)
					}
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

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

func (c *client) sendPNGFile() {
	err := qrcode.WriteFile(url+c.room.Id, qrcode.Medium, 256, c.room.Id+".png")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := os.Remove(c.room.Id + ".png")
		if err != nil {
			log.Fatal(err)
		}
	}()

	file, err := os.Open(c.room.Id + ".png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	err = c.socket.WriteMessage(websocket.BinaryMessage, fileBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *Room) ServeHTTP(ctx *gin.Context) {
	if r.Id == "" {
		log.Fatal("room id is required")
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	socket, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Fatal("ServeHTTP failed to upgrade:", err)
		return
	}

	socket.WriteMessage(websocket.TextMessage, []byte("oi"))
	client := &client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}

	r.Join <- client
	defer func() {
		r.Leave <- client
		fmt.Println("client left", client)
	}()

	go client.write()
	client.read()
}
