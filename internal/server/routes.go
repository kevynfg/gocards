package server

import (
	"fmt"
	"net/http"
	"sync"

	game "gocards/cmd/api/game"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var secretKey = []byte("fusrodah")

type MyCustomClaims struct {
	UserID string `json:"user_id"`
	jwt.MapClaims
}

type ServerIO struct{}

type CreateRoom struct {
	GameDuration  int    `json:"gameDuration"`
	RoundDuration int    `json:"roundDuration"`
	Topic         string `json:"topic"`
	MaxPlayers    int    `json:"maxPlayers"`
	RoomAdmin     string `json:"roomAdmin,omitempty"`
}

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

	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))
	r.GET("/health", s.healthHandler)
	r.GET("/ws/room/create", func(c *gin.Context) {
		// user_id := c.Query("user_id")
		// token := c.Query("token")
		// if user_id == "" || token == "" {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and token are required to join a room"})
		// 	return
		// }
		// fmt.Println("user_id:", user_id, "token:", token)
		roomID := uuid.New().String()
		fmt.Println("room id:", roomID)
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
