package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// const TimeLayoutStr = "2006-01-02 15:04:05"

func Stop(c *gin.Context) {
	// userId := strconv.Itoa(GetIdFromRequest(c))
	var params StopParam
	if err := c.ShouldBind(&params); err != nil {
		logrus.Debug("stop charging params not bind")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}
	if len(params.CarId) == 0 {
		logrus.Debug(params)
		logrus.Debug("wwwwww")
		SendBaseResponse(c, errno.ParamErr, nil)
		return
	}

	temp_car_id, _ := strconv.ParseInt(params.CarId, 10, 64)
	// tmp_id, _ := db.GetUserIDFromCarID(context.Background(), temp_car_id)
	// if userId != strconv.Itoa(tmp_id) {
	// 	logrus.Debug(params)
	// 	SendBaseResponse(c, errno.ParamErr, nil)
	// 	logrus.Debug("yyyyyyy")
	// 	SendBaseResponse(c, errno.ParamErr, nil)
	// 	return
	// }
	// 修改车辆状态
	car, _ := db.GetCarFromCarID(context.Background(), temp_car_id)
	car.IsCharge = false
	err := db.UpdateCar(context.Background(), car)
	if err != nil {
		logrus.Debug("update Cars failed")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}

	// 修改对应详单
	// bill_id, _ := strconv.ParseInt(params.BillId, 10, 64)
	bill_id := temp_car_id
	bill, _ := db.GetBillFromBillId(context.Background(), bill_id)

	// bill_id, bill_gen_time, pile_id, start_time, charge_type 在开始充电时就填写好了

	TimeNow := time.Now().Format(constants.TimeLayoutStr) // end_time
	loc, _ := time.LoadLocation("Local")
	start_time, _ := time.ParseInLocation(constants.TimeLayoutStr, bill.StartTime, loc)
	time_now, _ := time.ParseInLocation(constants.TimeLayoutStr, TimeNow, loc)
	dur := time_now.Sub(start_time).Nanoseconds() * constants.Scale // 实际差了多少ns
	ns, _ := time.ParseDuration("1ns")
	end_time := start_time.Add(ns * time.Duration(dur)) // 实际结束时间
	bill.EndTime = end_time.Format(constants.TimeLayoutStr)

	duration := end_time.Sub(start_time) // 充电持续时间

	bill.ChargeTime = duration.String() // charging_time 默认为ns

	power := 10
	if bill.ChargeType == constants.QuickCharge {
		power = 30
	}
	bill.ChargeQuantity = duration.Hours() * float64(power) // charging_quantity

	bill.ServiceFee = 0.8 * bill.ChargeQuantity
	bill.ChargeFee = CalChargeFee(bill.StartTime, bill.EndTime, power)
	bill.TotalFee = bill.ServiceFee + bill.ChargeFee

	err = db.UpdateBill(context.Background(), bill)
	if err != nil {
		logrus.Debug("update Cars failed")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}
	logrus.Debug(bill)
	scheduler.WhenChargingStop(bill.CarId, bill.PileId)

	sendStopResponse(c, errno.Success)
}

func CalChargeFee(start string, end string, power int) float64 {
	loc, _ := time.LoadLocation("Local")
	// start 是 2022-06-01 15:14:56 格式
	// start_time 是 2022-06-01 15:14:56 +0800 CST
	start_time, _ := time.ParseInLocation(constants.TimeLayoutStr, start, loc) // time.Time格式
	end_time, _ := time.ParseInLocation(constants.TimeLayoutStr, end, loc)
	arr_fee := [7]float64{0.4, 0.7, 1.0, 0.7, 1.0, 0.7, 0.4}
	arr_time := [7]int{7, 10, 15, 18, 21, 23, 24}
	fee := 0.0
	// 所有的时间差都小于24h，但存在跨天的情况
	if start_time.Hour() >= end_time.Hour() { // 跨天 eg.startH=23, endH=16
		// 先算start到当天24点的价格
		next_day := end[0:strings.Index(end, " ")] + " 00:00:00"
		s_index := GetIndex(start_time.Hour()) // eg.16对应arr_fee下标3
		fee1 := CalHelper(start, next_day, s_index, 6, power, arr_fee, arr_time)

		// 再算0点到end的价格
		e_index := GetIndex(end_time.Hour())
		fee2 := CalHelper(next_day, end, 0, e_index, power, arr_fee, arr_time)

		fee = fee1 + fee2
	} else { // 不跨天
		s_index := GetIndex(start_time.Hour())
		e_index := GetIndex(end_time.Hour())
		fee = CalHelper(start, end, s_index, e_index, power, arr_fee, arr_time)
	}
	return fee // 元
}

func GetIndex(hour int) int {
	if hour >= 0 && hour < 7 {
		return 0
	}
	if hour >= 7 && hour < 10 {
		return 1
	}
	if hour >= 10 && hour < 15 {
		return 2
	}
	if hour >= 15 && hour < 18 {
		return 3
	}
	if hour >= 18 && hour < 21 {
		return 4
	}
	if hour >= 21 && hour < 23 {
		return 5
	}
	if hour >= 23 && hour < 24 {
		return 6
	} else {
		return -1
	}
}

func CalHelper(start string, end string, s_index int, e_index int, power int, arr_f [7]float64, arr_t [7]int) float64 {
	loc, _ := time.LoadLocation("Local")
	// eg.start 2022-06-21 16:14:56; end 2022-06-22 7:05:32
	fee := 0.0
	for i := s_index; i <= e_index; i++ {
		if i == s_index {
			if i == e_index {
				s_time, _ := time.ParseInLocation(constants.TimeLayoutStr, start, loc)
				e_time, _ := time.ParseInLocation(constants.TimeLayoutStr, end, loc)
				dur := e_time.Sub(s_time).Hours()
				fee += dur * arr_f[i]
			} else {
				t := arr_t[i]
				mid := start[0:strings.Index(start, " ")] + " " + strconv.Itoa(t) + ":00:00"
				s_time, _ := time.ParseInLocation(constants.TimeLayoutStr, start, loc)
				m_time, _ := time.ParseInLocation(constants.TimeLayoutStr, mid, loc)
				dur := m_time.Sub(s_time).Hours()
				fee += dur * arr_f[i]
			}

		} else if i == e_index {
			t := arr_t[i-1]
			mid := end[0:strings.Index(end, " ")] + " " + strconv.Itoa(t) + ":00:00"
			e_time, _ := time.ParseInLocation(constants.TimeLayoutStr, end, loc)
			m_time, _ := time.ParseInLocation(constants.TimeLayoutStr, mid, loc)
			dur := e_time.Sub(m_time).Hours()
			fee += dur * arr_f[i]
		} else {
			fee += arr_f[i] * (float64(arr_t[i] - arr_t[i-1]))
		}
	}
	return float64(power) * fee
}

type StopParam struct {
	CarId string `json:"car_id"`
	// BillId string `json:"bill_id"`
}

type StopResponse struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func sendStopResponse(c *gin.Context, err error) {
	Err := errno.ConvertErr(err)
	c.JSON(http.StatusOK, StopResponse{
		StatusCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
	})
}
