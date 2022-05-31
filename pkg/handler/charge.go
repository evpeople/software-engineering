package handler

import (
	"net/http"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ChargingParam struct {
	ChargingType     int   `json:"charging_type"`
	ChargingQuantity int   `json:"charging_quantity"`
	CarId            int64 `json:"car_id"`
}

func Charge(c *gin.Context) {

	userId := int64(GetIdFromRequest(c))
	var params ChargingParam
	if err := c.ShouldBind(&params); err != nil {
		logrus.Debug("charging params not bind")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}
	num := scheduler.WhenCarComing(userId, params.CarId, params.ChargingType, params.ChargingQuantity)
	sendChargingResponse(c, errno.Success, chargingRespData{num > 0, num})
}

type ChargingResponse struct {
	Resp       bool   `json:"resp"`
	Num        int    `json:"num"`
	StautsCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

type chargingRespData struct {
	Resp bool
	Num  int
}

func sendChargingResponse(c *gin.Context, err error, data chargingRespData) {
	Err := errno.ConvertErr(err)
	c.JSON(http.StatusOK, ChargingResponse{
		StautsCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		Resp:       data.Resp,
		Num:        data.Num,
	})
}
