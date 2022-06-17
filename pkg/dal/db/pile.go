package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"gorm.io/gorm"
)

type PileInfo struct {
	gorm.Model
	PileID                int     `json:"pile_id"`
	IsWork                bool    `json:"is_work"`
	ChargingTotalCount    int     `json:"charging_total_count"`
	ChargingTotalTime     string  `json:"charging_total_time"`
	ChargingTotalQuantity float64 `json:"charging_total_quantity"`
}

func (u *PileInfo) TableName() string {
	return constants.PileTableName
}

func MGetPileID(ctx context.Context, pileID int64) (*PileInfo, error) {
	// res := make([]*User, 0)
	res := new(PileInfo)
	if err := DB.WithContext(ctx).Where("id = ?", pileID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func MGetAllPiles(ctx context.Context) ([]*PileInfo, error) {
	res := make([]*PileInfo, 0)
	if err := DB.WithContext(ctx).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func CreatePile(ctx context.Context, piles []*PileInfo) error {
	return DB.WithContext(ctx).Create(piles).Error
}

func UpdatePile(ctx context.Context, a_pile *PileInfo) error {
	return DB.WithContext(ctx).Updates(a_pile).Error
}
