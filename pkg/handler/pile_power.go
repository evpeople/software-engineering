package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//启动\关闭充电桩
type PowerResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	PileStatus bool   `json:"pile_status"`
}

func ResetPilePower(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logrus.Debug("Param got wrong", err)
		sendPowerResponse(c, errno.ConvertErr(err), false)
		return
	}
	ty, err := strconv.Atoi(c.Query("pile_type"))
	if err != nil {
		logrus.Debug("Query got wrong", err)
		sendPowerResponse(c, errno.ConvertErr(err), false)
	}
	//找到对应充电桩
	tarPile, err := db.MGetPileTag(context.Background(), int64(id), int64(ty))
	if err != nil {
		logrus.Debug("Get PileID wrong", err.Error())
		sendPowerResponse(c, errno.ConvertErr(err), false)
	}
	//logrus.Debug("Before: ", tarPile.IsWork)
	tarPile.IsWork = !tarPile.IsWork
	//logrus.Debug("After: ", tarPile.IsWork)
	err = db.UpdatePile(context.Background(), tarPile)
	if err != nil {
		logrus.Debug("**update pile status failed")
		sendPowerResponse(c, errno.ConvertErr(err), false)
	}

	scheduler.ResetPileState(tarPile.PileID)
	sendPowerResponse(c, errno.Success, tarPile.IsWork)
}

func sendPowerResponse(c *gin.Context, err error, status bool) {
	Err := errno.ConvertErr(err)
	c.JSON(http.StatusOK, PowerResp{
		StatusCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		PileStatus: status,
	})
}
