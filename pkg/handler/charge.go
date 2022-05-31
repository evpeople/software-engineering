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
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}

	ok := scheduler.WhenCarComing(userId, carId, params.ChargingType, params.ChargingQuantity)
	sendChargingResponse(c, errno.Success, ok)
}

type ChargingParam struct {
	ChargingType     int `json:"charging_type"`
	ChargingQuantity int `json:"charging_quantity"`
}

type ChargingResponse struct {
	Resp       bool   `json:"resp"`
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func sendChargingResponse(c *gin.Context, err error, ok bool) {
	Err := errno.ConvertErr(err)
	c.JSON(http.StatusOK, ChargingResponse{
		StatusCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		Resp:       ok,
	})
}
