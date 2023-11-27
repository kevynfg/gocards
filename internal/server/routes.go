package server

import (
	"net/http"

	game "gocards/cmd/api/game"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)
	room := game.NewRoom()
	r.GET("/room", room.ServeHTTP)
	go room.Run()
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
