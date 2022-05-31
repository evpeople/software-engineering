package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
		sendStopResponse(c, errno.ParamErr)
	}
	tmp_id, _ := db.GetUserIDFromCarID(context.Background(), params.CarId)
	if userId != tmp_id {
		logrus.Debug(params)
		sendStopResponse(c, errno.ParamErr)
	}
	carId := params.CarId
	// todo: change isCharging
	car, _ := db.GetCarFromCarID(context.Background(), carId)
	car.IsCharge = false
	sendStopResponse(c, errno.Success)
}

type StopParam struct {
	UserId string `json:"user_id"`
	CarId  string `json:"car_id"`
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
