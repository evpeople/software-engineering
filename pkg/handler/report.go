package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Report(c *gin.Context) {
	piles, err := db.MGetAllPiles(context.Background())
	if err != nil {
		logrus.Debug(err.Error())
		SendReportsResponse(c, errno.ConvertErr(err), nil)
	}
	len := len(piles)

	reportsInfoVar := make([]ReportInfo, len)
	for i := 0; i < len; i++ {
		pile := piles[i]
		reportsInfoVar[i].Time = time.Now().Format(constants.TimeLayoutStr)
		reportsInfoVar[i].PileId = pile.PileID
		reportsInfoVar[i].PileChargingTotalCount = pile.ChargingTotalCount
		reportsInfoVar[i].PileChargingTotalTime = pile.ChargingTotalTime
		reportsInfoVar[i].PileChargingTotalQuantity = pile.ChargingTotalQuantity
		reportsInfoVar[i].PileChargingTotalFee, err = db.GetChargingTotalFeeFromPileId(context.Background(), int64(pile.PileID))
		if err != nil {
			logrus.Debug(err.Error())
			SendReportsResponse(c, errno.ConvertErr(err), nil)
		}

		reportsInfoVar[i].PileServiceTotalFee, err = db.GetServiceTotalFeeFromPileId(context.Background(), int64(pile.PileID))
		if err != nil {
			logrus.Debug(err.Error())
			SendReportsResponse(c, errno.ConvertErr(err), nil)
		}

		reportsInfoVar[i].PileTotalFee = reportsInfoVar[i].PileChargingTotalFee + reportsInfoVar[i].PileServiceTotalFee
	}

	SendReportsResponse(c, errno.Success, reportsInfoVar)

}

type ReportInfo struct {
	Time                      string  `json:"time"`
	PileId                    int     `json:"pile_id"`
	PileChargingTotalCount    int     `json:"pile_charging_total_count"`
	PileChargingTotalTime     string  `json:"pile_charging_total_time"`
	PileChargingTotalQuantity float64 `json:"pile_charging_total_quantity"`
	PileChargingTotalFee      float64 `json:"pile_charging_total_fee"`
	PileServiceTotalFee       float64 `json:"pile_service_total_fee"`
	PileTotalFee              float64 `json:"pile_total_fee"`
}

type ReportsResp struct {
	StatusMsg  string       `json:"status_msg"`
	StatusCode int          `json:"status_code"`
	Reports    []ReportInfo `json:"reports"`
}

func SendReportsResponse(c *gin.Context, err error, data []ReportInfo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, CarsResp{
			StatusMsg:  Err.ErrMsg,
			StatusCode: Err.ErrCode,
		})
		return
	}
	c.JSON(http.StatusOK, ReportsResp{
		StatusMsg:  Err.ErrMsg,
		StatusCode: Err.ErrCode,
		Reports:    data,
	})
}
