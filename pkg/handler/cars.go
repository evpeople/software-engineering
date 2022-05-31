package handler

import (
	"net/http"
	"strconv"
	"time"

	// "time"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const TimeLayoutStr = "2006-01-02 15:04:05"

func Cars(c *gin.Context) {
	var carsInfoVar CarsInfo

	carsInfoVar.UserID = GetIdFromRequest(c)

	carsInfoVar.CarID = 4 //todo: add function to get car id
	carsInfoVar.CarCapacity, _ = strconv.Atoi(c.Query("car_capacity"))
	carsInfoVar.RequestedQuantity, _ = strconv.Atoi(c.Query("charging_quantity"))

	StartWaitingTime := c.Query("start_waiting_time")                    //todo: waiting define of 'start waiting time'
	loc, _ := time.LoadLocation("Local")                                 //获取本地时区
	t, err := time.ParseInLocation(TimeLayoutStr, StartWaitingTime, loc) //使用模板在对应时区转化为time.time类型
	if err != nil {
		logrus.Debug("start waiting time parse unsuccessful")
		sendRegisterResponse(c, errno.ConvertErr(err), nil)
		return
	}
	carsInfoVar.WaitingTime = time.Since(t).String()

	SendCarsResponse(c, errno.Success, &carsInfoVar)
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
