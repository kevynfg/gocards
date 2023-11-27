package server

import (
	"net/http"
	"sync"

	game "gocards/cmd/api/game"

	"github.com/gin-gonic/gin"
)

var rooms = make(map[string]*game.Room)
var lock = sync.RWMutex{}

func getRoom(roomID string) *game.Room {
	lock.RLock()
	defer lock.RUnlock()
	return rooms[roomID]
}

func createRoom(roomID string) *game.Room {
	lock.Lock()
	defer lock.Unlock()
	room := game.NewRoom()
	room.Id = roomID
	rooms[roomID] = room
	go room.Run()
	return room
}

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)
	r.GET("/room/:roomID", func(c *gin.Context) {
		if c.Param("roomID") == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required to join a room"})
			return
		}
		roomID := c.Param("roomID")
		room := getRoom(roomID)
		if room == nil {
			room = createRoom(roomID)
		}
		room.ServeHTTP(&gin.Context{Request: c.Request, Writer: c.Writer})
	})
	
	return r
}

	func (s *Server) HelloWorldHandler(c *gin.Context) {
		resp := make(map[string]string)
		resp["message"] = "Hello World"

		c.JSON(http.StatusOK, resp)
	}

	func (s *Server) healthHandler(c *gin.Context) {
		c.JSON(http.StatusOK, s.db.Health())
	}
