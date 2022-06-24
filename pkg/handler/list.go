package handler

import (
	"net/http"

	"github.com/evpeople/softEngineer/pkg/scheduler"
	"github.com/gin-gonic/gin"
)

type WaitingCar struct {
	CarID    int `json:"car_id"`
	CType    int `json:"ctype"`
	Quantity int `json:"quantity"`
}

func List(c *gin.Context) {
	waitingCar := []WaitingCar{}
	for i := scheduler.S.WaitingArea.Front(); i != nil; i = i.Next() {
		c := i.Value.(*scheduler.Car)
		waitingCar = append(waitingCar, WaitingCar{CarID: int(c.GetCarId()), Quantity: c.GetChargingQuantity(), CType: c.GetCType()})
	}
	c.JSON(http.StatusOK, waitingCar)
}
