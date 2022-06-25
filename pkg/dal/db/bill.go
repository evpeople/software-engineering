package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"gorm.io/gorm"
)

type Bill struct {
	gorm.Model
	CarId          int     `json:"car_id"`
	BillId         int     `json:"bill_id" gorm:"unique"`
	BillGenTime    string  `json:"bill_generate_time"`
	PileId         int     `json:"pile_id"`
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
	if err := DB.WithContext(ctx).Where("bill_id = ?", BillID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func GetBillFromId(ctx context.Context, ID int64) (*Bill, error) {
	res := new(Bill)
	if err := DB.WithContext(ctx).Where("id = ?", ID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func GetChargingBillFromPileId(ctx context.Context, pileId int64) (*Bill, error) {
	res := new(Bill)
	if err := DB.WithContext(ctx).Where("pile_id = ? and end_time is null and start_time is not null", pileId).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func GetChgSevTotalFeeFromPileId(ctx context.Context, pileId int64) (float64, float64, error) {
	res := make([]*Bill, 0)
	var chargeTotalFee float64 = 0
	var serviceTotalFee float64 = 0
	if err := DB.WithContext(ctx).Where("pile_id = ?", pileId).Find(&res).Error; err != nil {
		return -1, -1, err
	}
	for i := 0; i < len(res); i++ {
		chargeTotalFee += res[i].ChargeFee
		serviceTotalFee += res[i].ServiceFee
	}
	return chargeTotalFee, serviceTotalFee, nil
}

func UpdateBill(ctx context.Context, a_bill *Bill) error {
	return DB.WithContext(ctx).Updates(a_bill).Error
}

func CreateBill(ctx context.Context, bills []*Bill) error {
	return DB.WithContext(ctx).Create(bills).Error
}
