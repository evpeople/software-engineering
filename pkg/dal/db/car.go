package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"gorm.io/gorm"
)

type Car struct {
	gorm.Model
	UserName  string `json:"user_name" gorm:"unique"`
	Password  string `json:"password"`
	UserRefer int
	User      User `gorm:"foreignKey:UserRefer"`
}

func (u *Car) TableName() string {
	return constants.CarsTableName
}

// MGetUsers multiple get list of user info
func MGetCarFromUserID(ctx context.Context, userID int64) ([]*Car, error) {
	res := make([]*Car, 0)
	// res := new(Car)
	if err := DB.WithContext(ctx).Where("user_refer = ?", userID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func GetCarFromCarID(ctx context.Context, carID int64) (*Car, error) {
	// res := make([]*User, 0)
	res := new(Car)
	if err := DB.WithContext(ctx).Where("id = ?", carID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// CreateUser create user info
func CreateCar(ctx context.Context, cars []*Car) error {
	return DB.WithContext(ctx).Create(cars).Error
}

// QueryUserExist query list of user info
