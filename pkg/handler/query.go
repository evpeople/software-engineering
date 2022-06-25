package handler

import (
	"net/http"
	"strconv"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetUserQueue(c *gin.Context) {
	carId, err := strconv.Atoi(c.Query("car_id"))
	if err != nil {
		logrus.Debug(err)
		sendUserQueueResp(c, errno.ConvertErr(err), nil)
		return
	}
	var data UserQueueInfo
	data.QueueId, data.Num, data.Area = scheduler.GetQueueInfoByCarId(carId)
	sendUserQueueResp(c, nil, &data)
}

type UserQueueResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	QueueId    int    `json:"queue_id"`
	Num        int    `json:"num"`
	Area       int    `json:"area"`
}

type UserQueueInfo struct {
	QueueId int
	Num     int
	Area    int
}

func sendUserQueueResp(c *gin.Context, err error, data *UserQueueInfo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, UserQueueResp{
			StatusMsg:  Err.ErrMsg,
			StatusCode: Err.ErrCode,
			QueueId:    -1,
			Num:        -1,
			Area:       -1,
		})
		return
	}
	c.JSON(http.StatusOK, UserQueueResp{
		StatusMsg:  Err.ErrMsg,
		StatusCode: Err.ErrCode,
		QueueId:    data.QueueId,
		Num:        data.Num,
		Area:       data.Area,
	})
	return
}
