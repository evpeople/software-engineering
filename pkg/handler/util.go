package handler

import (
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/errno"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var AuthMiddleware *jwt.GinJWTMiddleware

type RegisterResponse struct {
	UserID     int    `json:"user_id"`
	Token      string `json:"token"`
	StautsCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func init() {
	AuthMiddleware, _ = jwt.New(&jwt.GinJWTMiddleware{
		Key:        []byte(constants.SecretKey),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			logrus.Debug("in Pay load", data)
			if v, ok := data.(int); ok {
				return jwt.MapClaims{
					"ID": v,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVar UserParam
			if err := c.ShouldBind(&loginVar); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			if len(loginVar.UserName) == 0 || len(loginVar.PassWord) == 0 {
				return "", jwt.ErrMissingLoginValues
			}
			id, err := CheckUser(loginVar)
			c.Set("userID", id)
			return id, err
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		LoginResponse: func(c *gin.Context, code int, message string, time time.Time) {
			c.JSON(http.StatusOK, RegisterResponse{
				Token:      message,
				UserID:     c.GetInt("userID"),
				StautsCode: code,
				StatusMsg:  errno.Success.ErrMsg,
			})
		},
	})
}
func SendRegisterResponse(c *gin.Context, err error, data *UserResp) {
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

type UserParam struct {
	UserName string `json:"username"`
	PassWord string `json:"password"`
}
type UserResp struct {
	UserID int
	Token  string
}

func GetIdFromRequest(c *gin.Context) int {
	return int(jwt.ExtractClaims(c)["ID"].(float64))
}
func SendBaseResponse(c *gin.Context, err error, data *UserResp) {
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
