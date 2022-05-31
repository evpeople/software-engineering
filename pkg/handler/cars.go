package handler

import (
	"net/http"

	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Cars(c *gin.Context) {
	var carsInfoVar CarsInfo

	if err := c.ShouldBind(&carsInfoVar); err != nil {
		logrus.Debug("not bind")

	}
}

type CarsParam struct {
	Token string `json:"token"`
}

type CarsInfo struct {
	UserID          int
	CarCapacity     int
	RequestQuantity int
	WaitingTime     string
}

type CarsResp struct {
	StatusMsg  string
	StautsCode int
	cars       CarsInfo
}

func SendCarsResponse(c *gin.Context, err error, data *CarsInfo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, CarsResp{
			StatusMsg:  Err.ErrMsg,
			StautsCode: Err.ErrCode,
		})
		return
	}
	c.JSON(http.StatusOK, CarsResp{
		StatusMsg:  Err.ErrMsg,
		StautsCode: Err.ErrCode,
		cars: CarsInfo{
			UserID:          data.UserID,
			CarCapacity:     data.CarCapacity,
			RequestQuantity: data.RequestQuantity,
			WaitingTime:     data.WaitingTime,
		},
	})
}
