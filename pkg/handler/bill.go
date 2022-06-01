package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetBill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logrus.Debug(err)
		sendBillResponse(c, errno.ConvertErr(err), nil)
	}
	// get bill from id
	bill, err := db.GetBillFromBillId(context.Background(), int64(id))
	if err != nil {
		logrus.Debug("Get Bill	wrong", err.Error())
		sendBillResponse(c, errno.ConvertErr(err), nil)
	}
	var billInfoVar BillInfo
	billInfoVar.BillId = bill.BillId
	billInfoVar.BillGenTime = bill.BillGenTime
	billInfoVar.PipeId = bill.PipeId
	billInfoVar.ChargeQuantity = bill.ChargeQuantity
	billInfoVar.ChargeType = bill.ChargeType
	billInfoVar.ChargeTime = bill.ChargeTime
	billInfoVar.StartTime = bill.StartTime
	billInfoVar.EndTime = bill.EndTime
	billInfoVar.ChargeFee = bill.ChargeFee
	billInfoVar.ServiceFee = bill.ServiceFee
	billInfoVar.TotalFee = bill.TotalFee

	sendBillResponse(c, errno.Success, &billInfoVar)
}

type BillInfo struct {
	BillId         int     `json:"bill_id" gorm:"unique"`
	BillGenTime    string  `json:"bill_generate_time"`
	PipeId         int     `json:"pipe_id"`
	ChargeQuantity float64 `json:"charging_quantity"`
	ChargeType     int     `json:"charging_type"`
	ChargeTime     string  `json:"charging_time"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	ChargeFee      float64 `json:"charging_fee"`
	ServiceFee     float64 `json:"service_fee"`
	TotalFee       float64 `json:"total_fee"`
}

type BillResp struct {
	StatusMsg  string     `json:"status_msg"`
	StatusCode int        `json:"status_code"`
	Bill       []BillInfo `json:"bill"`
}

func sendBillResponse(c *gin.Context, err error, data *BillInfo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, BillResp{
			StatusMsg:  Err.ErrMsg,
			StatusCode: Err.ErrCode,
		})
		return
	}
	c.JSON(http.StatusOK, BillResp{
		StatusMsg:  Err.ErrMsg,
		StatusCode: Err.ErrCode,
		Bill: []BillInfo{{
			BillId:         data.BillId,
			BillGenTime:    data.BillGenTime,
			PipeId:         data.PipeId,
			ChargeQuantity: data.ChargeQuantity,
			ChargeType:     data.ChargeType,
			ChargeTime:     data.ChargeTime,
			StartTime:      data.StartTime,
			EndTime:        data.EndTime,
			ChargeFee:      data.ChargeFee,
			ServiceFee:     data.ServiceFee,
			TotalFee:       data.TotalFee,
		}},
	})
}
