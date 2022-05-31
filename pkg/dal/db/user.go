package db

import (
	"context"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/errno"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserName string `json:"user_name" gorm:"unique"`
	Password string `json:"password"`
}

func (u *User) TableName() string {
	return constants.UserTableName
}

// MGetUsers multiple get list of user info
func MGetUser(ctx context.Context, userID int64) (*User, error) {
	// res := make([]*User, 0)
	res := new(User)
	if err := DB.WithContext(ctx).Where("id = ?", userID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// CreateUser create user info
func CreateUser(ctx context.Context, users []*User) error {
	return DB.WithContext(ctx).Create(users).Error
}

// QueryUserExist query list of user info
func QueryUserExist(ctx context.Context, username string) (*User, error) {
	// res := make([]*User, 0)
	res := new(User)
	// ans:=DB.First(res, "user_name = ?", username)
	if err := DB.First(res, "user_name = ?", username).Error; err != nil {
		//没有找到数据，可能返回的是 RecordNotExist
		return res, err
	}
	return res, nil
}
func CheckUser(ctx context.Context, username, password string) (int, error) {
	res := new(User)
	if err := DB.First(res, "user_name = ?", username).Error; err != nil {
		//没有找到数据，可能返回的是 RecordNotExist
		return -1, err
	}
	if res.Password != password {
		return -1, errno.LoginErr
	} else {
		return int(res.ID), nil
	}
}
