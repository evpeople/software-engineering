package handler

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"

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
		sendRegisterResponse(c, errno.ConvertErr(err), nil)
		return
	}

	if len(registerVar.UserName) == 0 || len(registerVar.PassWord) == 0 {
		logrus.Debug(registerVar)
		sendRegisterResponse(c, errno.ParamErr, nil)
		return
	}
	_, err := db.QueryUserExist(context.Background(), registerVar.UserName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		//需要的情况
		id, err := CreateUser(&db.User{UserName: registerVar.UserName, Password: registerVar.PassWord})
		if err != nil {
			logrus.Debug(err)
		}
		token, _, _ := AuthMiddleware.TokenGenerator(int(id))
		sendRegisterResponse(c, errno.Success, &UserResp{UserID: int(id), Token: token})
		return
	} else {
		sendRegisterResponse(c, errno.UserAlreadyExistErr, &UserResp{UserID: -1, Token: ""})
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
func sendRegisterResponse(c *gin.Context, err error, data *UserResp) {
	Err := errno.ConvertErr(err)
	if data == nil {
		c.JSON(http.StatusOK, RegisterResponse{
			StautsCode: Err.ErrCode,
			StatusMsg:  Err.ErrMsg,
		})
		return
	}
	c.JSON(http.StatusOK, RegisterResponse{
		StautsCode: Err.ErrCode,
		StatusMsg:  Err.ErrMsg,
		UserID:     data.UserID,
		Token:      data.Token,
	})
}

type RegisterResponse struct {
	UserID     int    `json:"user_id"`
	Token      string `json:"token"`
	StautsCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
