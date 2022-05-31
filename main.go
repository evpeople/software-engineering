package main

import (
	"net/http"

	"github.com/evpeople/softEngineer/pkg/dal"
	"github.com/evpeople/softEngineer/pkg/handler"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	scheduler.Init()
	dal.Init()
}
func main() {
	r := setupRouter()
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
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
	charge := v1.Group("/charge")
	charge.POST("/come", handler.Charge)
	admin := v1.Group("/admin")
	admin.GET("/cars", handler.Cars)
	admin.GET("/report", handler.Report)

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	//使用 下面的 test.Use()语句，包裹你所开发的api组，如此方能在 传入的 gin.Context上面，
	//通过 GetIdFromRequest方法获取用户的ID
	test := r.Group("/test")
	test.Use(handler.AuthMiddleware.MiddlewareFunc())
	test.Use(handler.Cars)
	test.GET("/ping", func(ctx *gin.Context) {
		id := handler.GetIdFromRequest(ctx)
		ctx.JSON(http.StatusOK, gin.H{
			"id": id,
		})

	})
	return r
}
