package handler

import (
	"context"
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
		SendCarsResponse(c, errno.PileNotExistErr, nil)
		return
	}

	var carsInfoVar []CarInfo
	var car *scheduler.Car

	// //! test code
	// pile.ChargeArea.PushBack(scheduler.NewCar(2, 1, 0, 0, 1000))
	// pile.ChargeArea.PushBack(scheduler.NewCar(2, 2, 0, 0, 1000))
	// pile.ChargeArea.PushBack(scheduler.NewCar(2, 3, 0, 0, 1050))

	if len := pile.ChargeArea.Len(); len <= 1 {
		// 队列中没有车或者只有一辆车在充电，没有等待车辆
		logrus.Debug(errno.NoWaitingCar.Error())
		SendCarsResponse(c, errno.NoWaitingCar, nil)
		return
	} else {
		// 有等待车辆，返回所有等待车辆的信息
		carsInfoVar = make([]CarInfo, len-1)

		// 计算当前正在充电的车，充完电所需的时间
		car = pile.ChargeArea.Front().Value.(*scheduler.Car)
		bill, err := db.GetChargingBillFromPileId(context.Background(), int64(pileId))
		if err != nil {
			logrus.Debug(err.Error())
			SendCarsResponse(c, errno.ConvertErr(err), nil)
			return
		}
		logrus.Debug("bill id: " + strconv.Itoa(bill.BillId))
		loc, _ := time.LoadLocation("Local")
		startTime, _ := time.ParseInLocation(constants.TimeLayoutStr, bill.StartTime, loc)
		logrus.Debug("startTime: " + startTime.Format(constants.TimeLayoutStr))
		totalTime, err := time.ParseDuration(strconv.FormatFloat(float64(car.GetChargingQuantity())/float64(pile.Power), 'f', 4, 64) + "h")
		logrus.Debug("totalTime: " + totalTime.String())
		if err != nil {
			logrus.Debug(err.Error())
			SendCarsResponse(c, errno.ConvertErr(err), nil)
			return
		}
		endTime := startTime.Add(totalTime)
		logrus.Debug("endTime: " + endTime.Format(constants.TimeLayoutStr))
		remainTime := time.Until(endTime)
		logrus.Debug("remainTime: " + remainTime.String())

		// 获取所有车辆的信息
		n := 0
		for i := pile.ChargeArea.Front().Next(); i != nil; i = i.Next() {
			car = i.Value.(*scheduler.Car)
			carsInfoVar[n].UserID = car.GetUserId()
			carsInfoVar[n].CarID = car.GetCarId()
			carsInfoVar[n].RequestedQuantity = float64(car.GetChargingQuantity())
			carsInfoVar[n].WaitingTime = remainTime.String()

			dbCar, err := db.GetCarFromCarID(c, carsInfoVar[n].CarID)
			if err != nil {
				logrus.Debug(err)
				SendCarsResponse(c, errno.ConvertErr(err), nil)
				return
			}
			carsInfoVar[n].CarCapacity = float64(dbCar.BatteryCap)

			// 更新 remain time
			newRemain := float64(car.GetChargingQuantity())/float64(pile.Power) + remainTime.Hours()
			remainTime, err = time.ParseDuration(strconv.FormatFloat(newRemain, 'f', 4, 64) + "h")
			if err != nil {
				logrus.Debug(err.Error())
				SendCarsResponse(c, errno.ConvertErr(err), nil)
				return
			}

			n++
		}
		SendCarsResponse(c, errno.Success, carsInfoVar)
	}
}

type CarInfo struct {
	UserID            int64   `json:"user_id"`
	CarID             int64   `json:"car_id"`
	CarCapacity       float64 `json:"car_capacity"`
	RequestedQuantity float64 `json:"requested_quantity"`
	WaitingTime       string  `json:"waiting_time"`
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
			CarsInfo:   []CarInfo{},
		})
		return
	}
	c.JSON(http.StatusOK, CarsResp{
		StatusMsg:  Err.ErrMsg,
		StatusCode: Err.ErrCode,
		CarsInfo:   data,
	})
}
