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
	tmp_car_id, _ := strconv.ParseInt(params.CarId, 10, 64)
	tmp_usr_id, _ := db.GetUserIDFromCarID(context.Background(), tmp_car_id)
	if userId != strconv.Itoa(tmp_usr_id) {
		logrus.Debug(params)
		SendBaseResponse(c, errno.ParamErr, nil)
	}
	car, _ := db.GetCarFromCarID(context.Background(), tmp_car_id)
	car.IsCharge = false
	sendStopResponse(c, errno.Success)
}

type StopParam struct {
	CarId string `json:"car_id"`
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
