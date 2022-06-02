package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// const TimeLayoutStr = "2006-01-02 15:04:05"

func Stop(c *gin.Context) {
	userId := strconv.Itoa(GetIdFromRequest(c))
	var params StopParam
	if err := c.ShouldBind(&params); err != nil {
		logrus.Debug("stop charging params not bind")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}
	if len(params.CarId) == 0 {
		logrus.Debug(params)
		SendBaseResponse(c, errno.ParamErr, nil)
	}

	temp_car_id, _ := strconv.ParseInt(params.CarId, 10, 64)
	tmp_id, _ := db.GetUserIDFromCarID(context.Background(), temp_car_id)
	if userId != strconv.Itoa(tmp_id) {
		logrus.Debug(params)
		SendBaseResponse(c, errno.ParamErr, nil)
	}
	// 修改车辆状态
	car, _ := db.GetCarFromCarID(context.Background(), temp_car_id)
	car.IsCharge = false
	err := db.UpdateCar(context.Background(), car)
	if err != nil {
		logrus.Debug("update Cars failed")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
	}

	// 修改对应详单
	bill_id, _ := strconv.ParseInt(params.BillId, 10, 64)
	bill, _ := db.GetBillFromBillId(context.Background(), bill_id)

	// bill_id, bill_gen_time, pipe_id, start_time, charge_type 在开始充电时就填写好了

	bill.EndTime = time.Now().Format(constants.TimeLayoutStr) // end_time

	loc, _ := time.LoadLocation("Local")
	start_time, _ := time.ParseInLocation(constants.TimeLayoutStr, bill.StartTime, loc)
	duration := time.Since(start_time) // 充电持续时间

	bill.ChargeTime = duration.String() // charging_time 默认为ns

	power := 10
	if bill.ChargeType == 0 {
		power = 30
	}
	bill.ChargeQuantity = duration.Hours() * float64(power) // charging_quantity

	bill.ServiceFee = 0.8 * bill.ChargeQuantity
	bill.ChargeFee = CalChargeFee(bill.StartTime, bill.EndTime, bill.ChargeQuantity)
	bill.TotalFee = bill.ServiceFee + bill.ChargeFee

	err = db.UpdateBill(context.Background(), bill)
	if err != nil {
		logrus.Debug("update Cars failed")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
	}

	sendStopResponse(c, errno.Success)
}

func CalChargeFee(start string, end string, quantity float64) float64 {
	// 根据开始充电和结束充电时间，区间计费
	// loc, _ := time.LoadLocation("Local")
	// start_time, _ := time.ParseInLocation(TimeLayoutStr, start, loc) // time.Time格式
	// end_time, _ := time.ParseInLocation(TimeLayoutStr, end, loc)
	// if start_time.Day() != end_time.Day() || start_time.Month() != end_time.Month() { // 跨天充电

	// } else {

	// }
	// todo: Calculate Function
	return 22.3 // 元
}

type StopParam struct {
	CarId  string `json:"car_id"`
	BillId string `json:"bill_id"`
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
