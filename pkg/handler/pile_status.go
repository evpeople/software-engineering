package handler

import (
	"context"
	"net/http"
	//"errors"

	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	//"gorm.io/gorm"
)

//查看所有充电桩状态
type PileResp struct {
	StatusCode int        `json:"status_code"`
	StatusMsg  string     `json:"status_msg"`
	Pile       []PileRepo `json:"pile"`
}

type PileRepo struct {
	PileID                int     `json:"pile_id"`
	IsWork                bool    `json:"is_work"`
	ChargingTotalCount    int     `json:"charging_total_count"`
	ChargingTotalTime     string  `json:"charging_total_time"`
	ChargingTotalQuantity float64 `json:"charging_total_quantity"`
}

func GetPileStatus(c *gin.Context) {
	piles, err := db.MGetAllPiles(context.Background())
	if err != nil {
		logrus.Debug(err.Error())
		SendReportsResponse(c, errno.ConvertErr(err), nil)
	}
	len := len(piles)

	PileStatusVar := make([]PileRepo, len)
	for i := 0; i < len; i++ {
		status := piles[i]
		if err != nil {
			logrus.Debug("**Get pile status failed", err.Error())
			sendPileResponse(c, errno.ConvertErr(err), nil)
		}
		PileStatusVar[i].PileID = status.PileID
		PileStatusVar[i].IsWork = status.IsWork
		PileStatusVar[i].ChargingTotalCount = status.ChargingTotalCount
		PileStatusVar[i].ChargingTotalTime = status.ChargingTotalTime
		PileStatusVar[i].ChargingTotalQuantity = status.ChargingTotalQuantity
	}

	sendPileResponse(c, errno.Success, PileStatusVar)
}

func sendPileResponse(c *gin.Context, err error, data []PileRepo) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, PileResp{
			StatusCode: Err.ErrCode,
			StatusMsg:  Err.ErrMsg,
		})
		return
	}
	c.JSON(http.StatusOK, PileResp{
		StatusCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		Pile:       data,
	})
}

/*
func CreatePile(req *db.PileInfo) (uint, error) {
	pile, err := db.QueryPileExist(context.Background(), req.PileTag, req.PileType)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		curPile := []*db.PileInfo{{
			PileID:req.PileID,
			PileType:req.PileType,
			PileTag:req.PileTag,
			IsWork:req.IsWork,
			ChargingTotalCount:req.ChargingTotalCount,
			ChargingTotalTime:req.ChargingTotalTime,
			ChargingTotalQuantity:req.ChargingTotalQuantity,
			Power:req.Power,
		}}
		err := db.CreatePile(context.Background(), curPile)
		return curPile[0].ID, err
	} else {
		return pile.ID, errno.UserAlreadyExistErr
	}
}
*/