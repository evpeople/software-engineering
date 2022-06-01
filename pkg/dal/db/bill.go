package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"gorm.io/gorm"
)

type Bill struct {
	gorm.Model
	BillId         int     `json:"bill_id" gorm:"unique"`
	BillGenTime    string  `json:"bill_generate_time"`
	PipeId         int     `json:"pipe_id"`
	ChargeQuantity float64 `json:"charging_quantity"`
	ChargeType     int     `json:"charging_type"`
	ChargeTime     string  `json:"charging_time"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	ChargeFee      float64 `json:"charging_fee"`
	ServiceFee     float64 `json:"service_fee"`
	TotalFee       float64 `json:"total_fee"`
}

func (u *Bill) TableName() string {
	return constants.BillTableName
}

func GetBillFromBillId(ctx context.Context, BillID int64) (*Bill, error) {
	res := new(Bill)
	if err := DB.WithContext(ctx).Where("id = ?", BillID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func CreateBill(ctx context.Context, bills []*Bill) error {
	return DB.WithContext(ctx).Create(bills).Error
}
