package scheduler

import (
	"container/list"
	"context"
	"sync"
	"time"
	"errors"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/evpeople/softEngineer/pkg/dal/db"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"github.com/evpeople/softEngineer/pkg/errno"
	"gorm.io/gorm"
)

// 充电桩状态的枚举类型
type PileStatus int

const (
	On        PileStatus = iota + 1 // EnumIndex = 1，充电桩开启
	Off                             // EnumIndex = 2，充电桩关闭
	Charging                        // EnumIndex = 3，充电桩正在充电
	Breakdown                       // EnumIndex = 4，充电桩故障
)

// 返回充电桩状态对应的字符串
func (status PileStatus) String() string {
	return [...]string{"On", "Off", "Charging", "Breakdown"}[status-1]
}

// 返回充电桩状态对应索引值
func (status PileStatus) EnumIndex() int {
	return int(status)
}

type Pile struct {
	PileId              int
	MaxWaitingNum       int
	Type                int64
	PileTag             int64
	Power               float32
	Status              PileStatus
	ChargeTotalCnt      int
	ChargeTotalQuantity float64

	Signal *semaphore.Weighted

	PileStartTime     int64
	PileStartTimeLock sync.Mutex

	emptyTimePredict int64
	WaitingArea      *list.List
	chargingCar      *Car
	CarsLock         sync.Mutex //lock for above 3
	// 充电时长（小时）=实际充电度数/充电功率(度/小时)，需要的时候再计算
}

// // 判断当前充电桩的队列是否满
// func (p *Pile) isFull() bool {
// 	return p.ChargeArea.Len() >= p.MaxWaitingNum
// }

// func (p *Pile) close() (bool, errno.ErrNo) {
// 	switch p.Status {
// 	case Off:
// 		return true, errno.Success
// 	case On:
// 		p.Status = Off
// 		return true, errno.Success
// 	case Breakdown:
// 		return false, errno.TurnOffBreakdownPileErr
// 	case Charging:
// 		// ? 需要考虑，充电中能否强制关机？能的话，需要添加后续处理；不能的话，需要返回错误信息
// 		// 此处暂时作为 充电中不能关机处理
// 		return false, errno.TurnOffChargingPileErr
// 	default:
// 		return true, errno.Success // 默认 Status 字段未初始化时，充电桩处于关闭状态
// 	}
// }

// func (p *Pile) open() (bool, errno.ErrNo) {
// 	switch p.Status {
// 	case Off:
// 		p.Status = On
// 		return true, errno.Success
// 	case On:
// 		return true, errno.Success
// 	case Breakdown:
// 		return false, errno.TurnOffBreakdownPileErr
// 	case Charging:
// 		// ? 需要考虑，充电中能否强制关机？能的话，需要添加后续处理；不能的话，需要返回错误信息
// 		// 此处暂时作为 充电中不能关机处理
// 		return false, errno.TurnOffChargingPileErr
// 	default:
// 		return true, errno.Success // 默认 Status 字段未初始化时，充电桩处于关闭状态
// 	}
// }

func NewPile(pileId int, maxWaitingNum int, pileType int64, pileTag int64, power float32, status PileStatus, siganl *semaphore.Weighted) *Pile {
	res := new(db.PileInfo)
	res.PileID = pileId
	res.PileType = int(pileType)
	res.PileTag = int(pileTag)
	if status == On {
		res.IsWork = true
	} else {
		res.IsWork = false
	}
	res.ChargingTotalCount = 0
	res.ChargingTotalTime = "0"
	res.ChargingTotalQuantity = 0
	res.Power = power
	err := CreatePile(res)
		if err != nil {
			logrus.Debug(err)
		}
	return &Pile{pileId, maxWaitingNum, pileType, pileTag, power, status, 0, 0,
		siganl, 0, sync.Mutex{}, time.Now().Unix(), list.New(), nil, sync.Mutex{}}
}

func CreatePile(req *db.PileInfo) (error) {
	err := db.QueryPileExist(context.Background(), req.PileTag, req.PileType)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		curPile := []*db.PileInfo{{
			PileID:req.PileID,
			PileType:req.PileType,
			PileTag:req.PileTag,
			IsWork:req.IsWork,
			ChargingTotalCount:req.ChargingTotalCount,
			ChargingTotalTime:req.ChargingTotalTime,
			ChargingTotalQuantity:req.ChargingTotalQuantity,
			Power:req.Power,
		}}
		err = db.CreatePile(context.Background(), curPile)
		return err
	} else {
		return errno.UserAlreadyExistErr
	}
}

func (p *Pile) isAlive() bool {
	p.PileStartTimeLock.Lock()
	ans := p.PileStartTime > 0
	p.PileStartTimeLock.Unlock()
	return ans
}

func (p *Pile) reStart() {
	p.PileStartTimeLock.Lock()
	p.PileStartTime = time.Now().Unix()
	p.PileStartTimeLock.Unlock()
}

func (p *Pile) shutdown() {
	p.PileStartTimeLock.Lock()
	p.PileStartTime = 0
	p.PileStartTimeLock.Unlock()
}

func (p *Pile) startTime() int64 {
	p.PileStartTimeLock.Lock()
	ans := p.PileStartTime
	p.PileStartTimeLock.Unlock()
	return ans
}

// func (p*Pile)TheChargingCar()Car{
// 	p.CarsLock.Lock()
// 	ans:=*p.chargingCar
// 	p.CarsLock.Unlock()
// 	return ans
// }

func (p *Pile) StartChargeNext() {
	go func() {

		p.CarsLock.Lock()
		next := p.WaitingArea.Front()
		if next == nil { //pile is empty
			p.CarsLock.Unlock()
			p.chargingCar = nil //charging nil
			return
		}
		car, ok := next.Value.(*Car)
		currentBill := &db.Bill{CarId: int(car.carId), BillId: int(car.carId), BillGenTime: time.Now().Format(constants.TimeLayoutStr), PileId: p.PileId, ChargeType: car.chargingType}
		if ok {
			currentBill.StartTime = time.Now().Format(constants.TimeLayoutStr) // start_time
			err := db.CreateBill(context.Background(), []*db.Bill{currentBill})
			if err != nil {
				logrus.Debug(err)
			}
			p.chargingCar = car //charging car
			p.WaitingArea.Remove(p.WaitingArea.Front())
		}
		p.CarsLock.Unlock()

		if ok {

			logrus.Info("pile ", p.PileId, " got car ", car.carId, "Start ")
			duration := float32(car.chargingQuantity) / p.Power
			startTime := time.Now().Unix()

			time.Sleep(time.Duration(duration) * time.Second)

			endTime := time.Now().Unix()

			t := p.startTime()
			if t < startTime && t > 0 {
				quantity := float64(p.Power) * float64((endTime-startTime)/1)
				p.ChargeTotalCnt++
				p.ChargeTotalQuantity += quantity
				p.CarsLock.Lock()

				//TODO: finish a charing: set the bill finish and other things here
				//TODO: when add codes notice that no blocking alows here
				p.CarsLock.Unlock()
				p.StartChargeNext()

			}
		}

	}()

}

func GetPileByTypeTag(pileType int64, pileTag int64) *Pile {
	var p *Pile
	var piles *list.List

	if pileType == ChargingType_Fast {
		piles = S.fastCharingPile
	} else if pileType == ChargingType_Trickle {
		piles = S.trickleChargingPile
	} else {
		return nil
	}

	for i := piles.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileTag == pileTag {
			return p
		}
	}
	return nil
}

func GetPileById(pileId int) *Pile {
	var p *Pile
	// 遍历慢充桩
	for i := S.trickleChargingPile.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileId == pileId {
			return p
		}
	}
	// 遍历快充桩
	for i := S.fastCharingPile.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileId == pileId {
			return p
		}
	}
	return nil
}
