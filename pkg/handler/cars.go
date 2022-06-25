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
	pileType, err := strconv.Atoi(c.Query("pile_type"))
	if err != nil {
		logrus.Debug(err)
		SendCarsResponse(c, errno.ConvertErr(err), nil)
		return
	}

	pileTag, err := strconv.Atoi(c.Query("pile_tag"))
	if err != nil {
		logrus.Debug(err)
		SendCarsResponse(c, errno.ConvertErr(err), nil)
		return
	}

	pile := scheduler.GetPileByTypeTag(int64(pileType), int64(pileTag))
	if pile == nil {
		logrus.Debug(errno.PileNotExistErr.Error())
		SendCarsResponse(c, errno.PileNotExistErr, nil)
		return
	}

	var carsInfoVar []CarInfo
	var car *scheduler.Car

	//! test code
	pile.WaitingArea.PushBack(scheduler.NewCar(2, 1, 0, 0, 1000))
	pile.WaitingArea.PushBack(scheduler.NewCar(2, 2, 0, 0, 1000))
	pile.WaitingArea.PushBack(scheduler.NewCar(2, 3, 0, 0, 1050))

	if len := pile.WaitingArea.Len(); len == 0 {
		// 队列中没有车
		logrus.Debug(errno.NoWaitingCar.Error())
		SendCarsResponse(c, errno.NoWaitingCar, nil)
		return
	} else {
		// 队列中有车辆，返回所有车辆的信息
		carsInfoVar = make([]CarInfo, len)

		// 设置当前正在充电的车的返回信息
		car = pile.WaitingArea.Front().Value.(*scheduler.Car)
		carsInfoVar[0].UserID = car.GetUserId()
		carsInfoVar[0].CarID = car.GetCarId()
		carsInfoVar[0].RequestedQuantity = float64(car.GetChargingQuantity())

		// 计算当前正在充电的车，充完电所需的时间
		bill, err := db.GetBillFromBillId(context.Background(), car.GetCarId()) // 由于测试数据中一个车只有一个订单，此处暂且这样写了。
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
		carsInfoVar[0].WaitingTime = remainTime.String()

		// 计算当前正在充电的车，已经充的电量和费用
		carsInfoVar[0].ChargedQuantity = float64(time.Since(startTime).Hours()) * float64(pile.Power)
		chargedFee := CalChargeFee(bill.StartTime, time.Now().Format(constants.TimeLayoutStr), int(pile.Power))
		carsInfoVar[0].CurrentFee = chargedFee + 0.8*carsInfoVar[0].ChargedQuantity

		// 获取后续等待车辆的信息
		n := 1
		for i := pile.WaitingArea.Front().Next(); i != nil; i = i.Next() {
			car = i.Value.(*scheduler.Car)
			carsInfoVar[n].UserID = car.GetUserId()
			carsInfoVar[n].CarID = car.GetCarId()
			carsInfoVar[n].RequestedQuantity = float64(car.GetChargingQuantity())
			carsInfoVar[n].WaitingTime = remainTime.String()
			carsInfoVar[n].ChargedQuantity = 0
			carsInfoVar[n].CurrentFee = 0

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
	ChargedQuantity   float64 `json:"charged_quantity"`
	CurrentFee        float64 `json:"current_fee"`
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
