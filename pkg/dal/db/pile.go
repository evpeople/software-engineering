package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PileInfo struct {
	gorm.Model
	PileID                int     `json:"pile_id"`
	PileType              int     `json:"pile_type"`
	PileTag               int     `json:"pile_tag"`
	IsWork                bool    `json:"is_work"`
	ChargingTotalCount    int     `json:"charging_total_count"`
	ChargingTotalTime     string  `json:"charging_total_time"`
	ChargingTotalQuantity float64 `json:"charging_total_quantity"`
	Power                 float32 `json:"power"`
}

func (u *PileInfo) TableName() string {
	return constants.PileTableName
}

func MGetPileID(ctx context.Context, pileID int64) (*PileInfo, error) {
	// res := make([]*User, 0)
	res := new(PileInfo)
	if err := DB.WithContext(ctx).Where("pile_id = ?", pileID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func MGetPileTag(ctx context.Context, pileTag int64, pileType int64) (*PileInfo, error) {
	// res := make([]*User, 0)
	res := new(PileInfo)
	if err := DB.WithContext(ctx).Where("pile_tag = ? AND pile_type = ?", pileTag, pileType).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func MGetAllPiles(ctx context.Context) ([]*PileInfo, error) {
	res := make([]*PileInfo, 0)
	if err := DB.WithContext(ctx).Find(&res).Error; err != nil {
		logrus.Debug("**Get all piles failed", err.Error())
		return nil, err
	}
	return res, nil
}

func CreatePile(ctx context.Context, piles []*PileInfo) error {
	return DB.WithContext(ctx).Create(piles).Error
}

func UpdatePile(ctx context.Context, a_pile *PileInfo) error {
	//logrus.Debug("In fact: ", a_pile.IsWork)
	return DB.WithContext(ctx).Select("is_work", "charging_total_count", "charging_total_time",
		"charging_total_quantity", "updated_at").Where("pile_id = ?", a_pile.PileID).Updates(a_pile).Error
}

func QueryPileExist(ctx context.Context, pileTag int, pileType int) (error) {
	res := new(PileInfo)
	if err := DB.First(res, "pile_tag = ? AND pile_type = ?", pileTag, pileType).Error; err != nil {
		//没有找到数据，可能返回的是 RecordNotExist
		return err
	}
	return nil
}
