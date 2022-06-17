package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetCarsInfo(c *gin.Context) {
	pileId, err := strconv.Atoi(c.Query("pile_id"))
	if err != nil {
		logrus.Debug(err)
		SendCarsResponse(c, errno.ConvertErr(err), nil)
		return
	}

	pile := scheduler.GetPileById(pileId)
	if pile == nil {
		logrus.Debug(errno.PileNotExistErr.Error())
		sendCarResponse(c, errno.PileNotExistErr, nil)
		return
	}

	var carsInfoVar []CarInfo
	var car *scheduler.Car

	if len := pile.ChargeArea.Len(); len <= 1 {
		// 队列中只有一辆车在充电或者没有车，没有等待车辆
		logrus.Debug("No waiting cars.")
		sendCarResponse(c, errno.Success, nil)
		return
	} else {
		// 有等待车辆，返回所有等待车辆的信息
		carsInfoVar = make([]CarInfo, len-1)
		// 计算当前正在充电的车，充完电所需的时间
		car = pile.ChargeArea.Front().Value.(*scheduler.Car)
		waitingHours := float32(car.GetChargingQuantity()) / pile.Power

		n := 0
		for i := pile.ChargeArea.Front().Next(); i != nil; i = i.Next() {
			car = i.Value.(*scheduler.Car)
			carsInfoVar[n].UserID = car.GetUserId()
			carsInfoVar[n].CarID = car.GetCarId()
			carsInfoVar[n].RequestedQuantity = car.GetChargingQuantity()
			// carsInfoVar[i].WaitingTime =

			dbCar, err := db.GetCarFromCarID(c, carsInfoVar[n].CarID)
			if err != nil {
				logrus.Debug(err)
				SendCarsResponse(c, errno.ConvertErr(err), nil)
				return
			}
			carsInfoVar[n].CarCapacity = dbCar.BatteryCap
			n++
		}

		// carsInfoVar.UserID = GetIdFromRequest(c)

		// carsInfoVar.CarID = 4 //todo: add function to get car id
		// carsInfoVar.CarCapacity, _ = strconv.Atoi(c.Query("car_capacity"))
		// carsInfoVar.RequestedQuantity, _ = strconv.Atoi(c.Query("charging_quantity"))

		StartWaitingTime := c.Query("start_waiting_time")                              //todo: waiting define of 'start waiting time'
		loc, _ := time.LoadLocation("Local")                                           //获取本地时区
		t, err := time.ParseInLocation(constants.TimeLayoutStr, StartWaitingTime, loc) //使用模板在对应时区转化为time.time类型
		if err != nil {
			logrus.Debug("start waiting time parse unsuccessful")
			SendCarsResponse(c, errno.ConvertErr(err), nil)
			return
		}
		// carsInfoVar.WaitingTime = time.Since(t).String()

		SendCarsResponse(c, errno.Success, carsInfoVar)

	}
}

type CarInfo struct {
	UserID            int64  `json:"user_id"`
	CarID             int64  `json:"car_id"`
	CarCapacity       int    `json:"car_capacity"`
	RequestedQuantity int    `json:"requested_quantity"`
	WaitingTime       string `json:"waiting_time"`
}

type CarsResp struct {
	StatusMsg  string    `json:"status_msg"`
	StatusCode int       `json:"status_code"`
	CarsInfo   []CarInfo `json:"cars_info"`
}

func SendCarsResponse(c *gin.Context, err error, data []CarInfo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, CarsResp{
			StatusMsg:  Err.ErrMsg,
			StatusCode: Err.ErrCode,
		})
		return
	}
	c.JSON(http.StatusOK, CarsResp{
		StatusMsg:  Err.ErrMsg,
		StatusCode: Err.ErrCode,
		CarsInfo:   data,
	})
}
