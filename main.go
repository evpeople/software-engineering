package main

import "github.com/gin-gonic/gin"

func main() {
	r:=setupRouter()
	r.Run()
}
func setupRouter() *gin.Engine {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.String(200, "pong")
    })
    return r
}
