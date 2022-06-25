package handler

import (
	"net/http"
	"strconv"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetWaitingNums(c *gin.Context) {
	queryType, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logrus.Debug(err)
		sendWaitingResponse(c, err, -1)
	}

	num := scheduler.GetAllWaiting(queryType)
	sendWaitingResponse(c, nil, num)
}

type WaitingResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	Num        int    `json:"num"`
}

func sendWaitingResponse(c *gin.Context, err error, num int) {
	Err := errno.ConvertErr(err)
	if num == -1 {
		c.JSON(http.StatusOK, WaitingResp{
			StatusMsg:  Err.ErrMsg,
			StatusCode: Err.ErrCode,
			Num:        num,
		})
		return
	}
	c.JSON(http.StatusOK, WaitingResp{
		StatusMsg:  Err.ErrMsg,
		StatusCode: Err.ErrCode,
		Num:        num,
	})
}
