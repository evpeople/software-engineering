package main

import (
	"github.com/evpeople/softEngineer/pkg/dal"
	"github.com/evpeople/softEngineer/pkg/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	dal.Init()
}
func main() {
	r := setupRouter()
	r.Run(":8080")
}
func setupRouter() *gin.Engine {
	r := gin.New()
	// logger := logrus.New()

	// //设置输出
	// // logger.Out = src

	// //设置日志级别
	// logger.SetLevel(logrus.DebugLevel)

	// //设置日志格式
	// logger.SetFormatter(&logrus.TextFormatter{
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// })
	logrus.SetLevel(logrus.DebugLevel)
	v1 := r.Group("/v1")
	user := v1.Group("/user")
	user.POST("/register", handler.Register)
	user.POST("/login", handler.AuthMiddleware.LoginHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return r
}
