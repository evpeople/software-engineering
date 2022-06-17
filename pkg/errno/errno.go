package errno

import (
	"errors"
	"fmt"
)

const (
	SuccessCode                 = 0
	ServiceErrCode              = 10001
	ParamErrCode                = 10002
	LoginErrCode                = 10003
	UserNotExistErrCode         = 10004
	UserAlreadyExistErrCode     = 10005
	TurnOffBreakdownPileErrCode = 10006
	TurnOffChargingPileErrCode  = 10007
	TurnOnBreakdownPileErrCode  = 10008
	PileNotExistErrCode         = 10009
)

type ErrNo struct {
	ErrCode int
	ErrMsg  string
}

func (e ErrNo) Error() string {
	return fmt.Sprintf("err_code=%d, err_msg=%s", e.ErrCode, e.ErrMsg)
}

func NewErrNo(code int, msg string) ErrNo {
	return ErrNo{code, msg}
}

func (e ErrNo) WithMessage(msg string) ErrNo {
	e.ErrMsg = msg
	return e
}

var (
	Success                 = NewErrNo(SuccessCode, "Success")
	ServiceErr              = NewErrNo(ServiceErrCode, "Service is unable to start successfully")
	ParamErr                = NewErrNo(ParamErrCode, "Wrong Parameter has been given")
	LoginErr                = NewErrNo(LoginErrCode, "Wrong username or password")
	UserNotExistErr         = NewErrNo(UserNotExistErrCode, "User does not exists")
	UserAlreadyExistErr     = NewErrNo(UserAlreadyExistErrCode, "User already exists")
	TurnOffBreakdownPileErr = NewErrNo(TurnOffBreakdownPileErrCode, "Can't turn off a broken-down charging pile")
	TurnOffChargingPileErr  = NewErrNo(TurnOffChargingPileErrCode, "Can't turn off a charging pile")
	TurnOnBreakdownPileErr  = NewErrNo(TurnOnBreakdownPileErrCode, "Can't turn on a broken-down charging pile")
	PileNotExistErr         = NewErrNo(PileNotExistErrCode, "Pile does not exists.")
)

// ConvertErr convert error to Errno
func ConvertErr(err error) ErrNo {
	if err == nil {
		return Success
	}
	Err := ErrNo{}
	if errors.As(err, &Err) {
		return Err
	}

	s := ServiceErr
	s.ErrMsg = err.Error()
	return s
}
