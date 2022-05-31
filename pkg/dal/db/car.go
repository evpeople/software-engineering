package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"gorm.io/gorm"
)

type Car struct {
	gorm.Model
	BatteryCap int
	UserRefer  int
	User       User `gorm:"foreignKey:UserRefer"`
	IsCharge   bool
}

func (u *Car) TableName() string {
	return constants.CarsTableName
}

// MGetCars multiple get list of Car info
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

func GetUserIDFromCarID(ctx context.Context, carID int64) (int, error) {
	// res := make([]*User, 0)
	res := new(Car)
	if err := DB.WithContext(ctx).Where("id = ?", carID).Find(&res).Error; err != nil {
		return -1, err
	}
	return res.UserRefer, nil
}

// CreateCar create Car info
func CreateCar(ctx context.Context, cars []*Car) error {
	return DB.WithContext(ctx).Create(cars).Error
}
func IsCharging(carID int) (bool, error) {
	car, err := GetCarFromCarID(context.Background(), int64(carID))
	if err != nil {
		return false, err
	}
	return car.IsCharge, nil
}
func NewCar(batteryCap int, userID int) (*Car, error) {
	user, err := MGetUser(context.Background(), int64(userID))
	if err != nil {
		return nil, err
	}
	return &Car{
		BatteryCap: batteryCap,
		User:       *user,
		UserRefer:  userID,
	}, nil
}

// QueryUserExist query list of user info
