package handler

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"

	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	var registerVar UserParam

	if err := c.ShouldBind(&registerVar); err != nil {
		logrus.Debug("not bind")
		SendRegisterResponse(c, errno.ConvertErr(err), nil)
		return
	}

	if len(registerVar.UserName) == 0 || len(registerVar.PassWord) == 0 {
		logrus.Debug(registerVar)
		SendRegisterResponse(c, errno.ParamErr, nil)
		return
	}
	_, err := db.QueryUserExist(context.Background(), registerVar.UserName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		//需要的情况
		id, err := CreateUser(&db.User{UserName: registerVar.UserName, Password: registerVar.PassWord})
		if err != nil {
			logrus.Debug(err)
		}
		token, _, _ := AuthMiddleware.TokenGenerator(id)
		SendRegisterResponse(c, errno.Success, &UserResp{UserID: int(id), Token: token})
		return
	} else {
		SendRegisterResponse(c, errno.UserAlreadyExistErr, &UserResp{UserID: -1, Token: ""})
	}

}
func CreateUser(req *db.User) (uint, error) {
	users, err := db.QueryUserExist(context.Background(), req.UserName)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		h := md5.New()
		if _, err = io.WriteString(h, req.Password); err != nil {
			return 0, err
		}
		passWord := fmt.Sprintf("%x", h.Sum(nil))
		ur := []*db.User{{
			UserName: req.UserName,
			Password: passWord,
		}}
		err := db.CreateUser(context.Background(), ur)
		return ur[0].ID, err
	} else {
		return users.ID, errno.UserAlreadyExistErr
	}
}
func CheckUser(loginVar UserParam) (int, error) {
	h := md5.New()
	if _, err := io.WriteString(h, loginVar.PassWord); err != nil {
		return 0, err
	}
	password := fmt.Sprintf("%x", h.Sum(nil))
	return db.CheckUser(context.Background(), loginVar.UserName, password)
}
