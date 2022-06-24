package handler

import (
	"net/http"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type QueryParam struct {
	ChargingType int `json:"charging_type"`
}

func Query(c *gin.Context) {

	// userId := int64(GetIdFromRequest(c))
	var params QueryParam
	if err := c.ShouldBind(&params); err != nil {
		logrus.Debug("charging params not bind")
		SendBaseResponse(c, errno.ConvertErr(err), nil)
		return
	}
	num := 1
	sendQueryResponse(c, errno.Success, QueryRespData{num})
}

type QueryResponse struct {
	Num int `json:"num"`

	StatusCode int `json:"status_code"`

	StatusMsg string `json:"status_msg"`
}

type QueryRespData struct {
	Num int
}

func sendQueryResponse(c *gin.Context, err error, data QueryRespData) {
	Err := errno.ConvertErr(err)
	c.JSON(http.StatusOK, QueryResponse{
		StatusCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		Num:        data.Num,
	})
}
