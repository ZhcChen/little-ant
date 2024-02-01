package main

import "github.com/gin-gonic/gin"

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", Ping)
		v1.GET("/ws", Ws)
	}

	r.Run(":26888")
}
