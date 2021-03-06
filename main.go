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
	dal.Init()
	scheduler.Init()
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

	//使用 下面的 api组.Use()语句，包裹 middlewarefunc，如此方能在传入的 gin.Context上面，
	//通过 GetIdFromRequest 方法获取用户的 ID
	test := r.Group("/test")
	test.Use(handler.AuthMiddleware.MiddlewareFunc())
	test.GET("/ping", func(ctx *gin.Context) {
		id := handler.GetIdFromRequest(ctx)
		ctx.JSON(http.StatusOK, gin.H{
			"id": id,
		})

	})

	v1 := r.Group("/v1")

	user := v1.Group("/user")
	user.POST("/register", handler.Register)
	user.POST("/login", handler.AuthMiddleware.LoginHandler)

	charge := v1.Group("/charge")
	charge.Use(handler.AuthMiddleware.MiddlewareFunc())
	charge.POST("/come", handler.Charge)
	charge.POST("/stop", handler.Stop)
	charge.GET("/list", handler.List)
	charge.GET("/:id", handler.GetBill)

	car := v1.Group("/car")
	car.Use(handler.AuthMiddleware.MiddlewareFunc())
	car.GET("/:id", handler.GetCarFromCarID)
	car.POST("", handler.AddCar)

	admin := v1.Group("/admin")
	admin.Use(handler.AuthMiddleware.MiddlewareFunc())
	admin.GET("/cars", handler.GetCarsInfo)
	admin.GET("/report", handler.Report)
	admin.POST("/pile/:id", handler.ResetPilePower)
	admin.GET("/status/pipe", handler.GetPileStatus)
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return r
}
