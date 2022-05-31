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

type CarResp struct {
	UserID int `json:"user_id"`
	Cap    int `json:"cap"`
	ID     int `json:"car_id"`
}

func AddCar(c *gin.Context) {
	// var carP CarParam
	// if err := c.ShouldBind(&carP); err != nil {
	// 	logrus.Debug("not bind AddCar")
	// 	sendCarResponse(c, errno.ConvertErr(err), nil)
	// 	return
	// }
	// logrus.Debug(carP)
	cap, err := strconv.Atoi(c.Query("cap"))
	if err != nil {
		logrus.Debug(err)
		sendCarResponse(c, errno.ConvertErr(err), nil)
	}
	car, err := db.NewCar(cap, GetIdFromRequest(c))
	if err != nil {
		logrus.Debug("New Car  wrong", err.Error())
		sendCarResponse(c, errno.ConvertErr(err), nil)

	}
	db.CreateCar(context.Background(), []*db.Car{car})
	sendCarResponse(c, nil, &CarResp{car.UserRefer, car.BatteryCap, int(car.ID)})
}
func GetCarFromCarID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logrus.Debug(err)
		sendCarResponse(c, errno.ConvertErr(err), nil)
	}
	car, err := db.GetCarFromCarID(context.Background(), int64(id))
	if err != nil {
		logrus.Debug("Get Car  wrong", err.Error())
		sendCarResponse(c, errno.ConvertErr(err), nil)
	}
	sendCarResponse(c, nil, &CarResp{car.UserRefer, car.BatteryCap, int(car.ID)})
}
func sendCarResponse(c *gin.Context, err error, data *CarResp) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, RegisterResponse{
			StautsCode: Err.ErrCode,
			StatusMsg:  Err.ErrMsg,
		})
		return
	}
	c.JSON(http.StatusOK, CarResponse{
		StautsCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		UserID:     data.UserID,
		CarID:      data.ID,
		Cap:        data.Cap,
	})
}

type CarResponse struct {
	UserID     int    `json:"user_id"`
	CarID      int    `json:"car_id"`
	Cap        int    `json:"cap"`
	StautsCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
