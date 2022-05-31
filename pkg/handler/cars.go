package handler

import (
	"net/http"
	// "time"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const TimeLayoutStr = "2006-01-02 15:04:05"

func Cars(c *gin.Context) {
	userId := GetIdFromRequest(c)

	var carsVar CarsParam
	if err := c.ShouldBind(&carsVar); err != nil {
		logrus.Debug("cars parameters failed to bind.")
		sendRegisterResponse(c, errno.ConvertErr(err), nil)
		return
	}

	//todo: 等待 car 的定义完善——进入等待区的时间
	// loc, _ := time.LoadLocation("Local")    //获取本地时区
	// t, err := time.ParseInLocation(TimeLayoutStr, carsVar.startWaitingTime, loc) //使用模板在对应时区转化为time.time类型
	// if err != nil {
	// 	logrus.Debug("start waiting time parse unsuccessful")
	// 	sendRegisterResponse(c, errno.ConvertErr(err), nil)
	// 	return
	// }

	SendCarsResponse(c, errno.Success, &CarsInfo{
		UserID:          userId,
		CarCapacity:     carsVar.CarCapacity,
		RequestQuantity: carsVar.RequestQuantity,
		// WaitingTime: time.Now().Sub(t).String(),
		WaitingTime: "1h20m10s",
	})
}

type CarsParam struct {
	CarCapacity     int
	RequestQuantity int
	// startWaitingTime     string
}

type CarsInfo struct {
	UserID          int
	CarCapacity     int
	RequestQuantity int
	WaitingTime     string
}

type CarsResp struct {
	StatusMsg  string
	StautsCode int
	cars       CarsInfo
}

func SendCarsResponse(c *gin.Context, err error, data *CarsInfo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, CarsResp{
			StatusMsg:  Err.ErrMsg,
			StautsCode: Err.ErrCode,
		})
		return
	}
	c.JSON(http.StatusOK, CarsResp{
		StatusMsg:  Err.ErrMsg,
		StautsCode: Err.ErrCode,
		cars: CarsInfo{
			UserID:          data.UserID,
			CarCapacity:     data.CarCapacity,
			RequestQuantity: data.RequestQuantity,
			WaitingTime:     data.WaitingTime,
		},
	})
}
