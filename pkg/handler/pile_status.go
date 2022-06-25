package handler

import (
	"context"
	"net/http"
	//"crypto/md5"

	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

//todo
/*
func CreatePile(req *db.PileInfo) (uint, error) {
	users, err := db.QueryPileExist(context.Background(), req.UserName)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		h := md5.New()
		if _, err = io.WriteString(h, req.Password); err != nil {
			return 0, err
		}
		passWord := fmt.Sprintf("%x", h.Sum(nil))
		ur := []*db.PileInfo{{
			UserName: req.UserName,
			Password: passWord,
		}}
		err := db.CreatePile(context.Background(), ur)
		return ur[0].ID, err
	} else {
		return users.ID, errno.UserAlreadyExistErr
	}
}
*/