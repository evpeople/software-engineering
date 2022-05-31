package handler

import (
	"net/http"
	"strconv"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Charge(c *gin.Context) {
	userId := strconv.Itoa(GetIdFromRequest(c))
	carId := "0"

	var params ChargingParam
	if err := c.ShouldBind(&params); err != nil {
		logrus.Debug("charging params not bind")
		SendRegisterResponse(c, errno.ConvertErr(err), nil)
		return
	}

	ok := scheduler.WhenCarComing(userId, carId, params.chargingType, params.chargingQuantity)
	sendChargingResponse(c, errno.Success, ok)
}

type ChargingParam struct {
	chargingType     int `json:"charging_type"`
	chargingQuantity int `json:"charging_quantity"`
}

type ChargingResponse struct {
	Resp       bool   `json:"resp"`
	StautsCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func sendChargingResponse(c *gin.Context, err error, ok bool) {
	Err := errno.ConvertErr(err)
	c.JSON(http.StatusOK, ChargingResponse{
		StautsCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		Resp:       ok,
	})
}
