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

	//todo: add function to get car id

	var carsVar CarsParam
	if err := c.ShouldBind(&carsVar); err != nil {
		logrus.Debug("cars parameters failed to bind.")
		sendRegisterResponse(c, errno.ConvertErr(err), nil)
		return
	}

	//todo: waiting define of 'start waiting time'
	// loc, _ := time.LoadLocation("Local")    //获取本地时区
	// t, err := time.ParseInLocation(TimeLayoutStr, carsVar.StartWaitingTime, loc) //使用模板在对应时区转化为time.time类型
	// if err != nil {
	// 	logrus.Debug("start waiting time parse unsuccessful")
	// 	sendRegisterResponse(c, errno.ConvertErr(err), nil)
	// 	return
	// }

	SendCarsResponse(c, errno.Success, &CarsInfo{
		UserID: userId,
		//todo CarID: carID,
		CarID:             4,
		CarCapacity:       carsVar.CarCapacity,
		RequestedQuantity: carsVar.ChargingQuantity,
		//todo WaitingTime: time.Now().Sub(t).String(),
		WaitingTime: carsVar.StartWaitingTime,
	})
}

type CarsParam struct {
	CarCapacity      int    `json:"car_capacity"`
	ChargingQuantity int    `json:"charging_quantity"`
	StartWaitingTime string `json:"start_waiting_time"`
}

type CarsInfo struct {
	UserID            int    `json:"user_id"`
	CarID             int    `json:"car_id"`
	CarCapacity       int    `json:"car_capacity"`
	RequestedQuantity int    `json:"requested_quantity"`
	WaitingTime       string `json:"waiting_time"`
}

type CarsResp struct {
	StatusMsg  string   `json:"status_msg"`
	StautsCode int      `json:"status_code"`
	Cars       CarsInfo `json:"cars"`
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
		Cars: CarsInfo{
			UserID:            data.UserID,
			CarID:             data.CarID,
			CarCapacity:       data.CarCapacity,
			RequestedQuantity: data.RequestedQuantity,
			WaitingTime:       data.WaitingTime,
		},
	})
}
