package scheduler

import (
	"container/list"
	"context"
	"strconv"
	"strings"
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
	MUL                  = constants.Scale
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
		siganl, time.Now().Unix(), sync.Mutex{}, time.Now().Unix(), list.New(), nil, sync.Mutex{}}
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
		logrus.Info("pile", p.PileId, ": Charge next")
		p.CarsLock.Lock()
		next := p.WaitingArea.Front()
		if next == nil { //pile is empty
			p.chargingCar = nil //charging nil
			logrus.Debug("pile ", p.PileId, " temp sleep")
			p.CarsLock.Unlock()
			return
		}
		car, ok := next.Value.(*Car)

		if ok {
			//TODO:car start charging here
			currentBill := &db.Bill{CarId: int(car.carId), BillId: int(car.carId), BillGenTime: time.Now().Format(constants.TimeLayoutStr), PileId: p.PileId, ChargeType: car.chargingType}

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
			duration := p.chargeTime(car.chargingQuantity)

			logrus.Info(time.Now(), "--", time.Now().Unix(), "pile ", p.PileId, " got car ", car.carId, "Start and will charge for ", duration, "ms.")
			startTime := time.Now().Unix()

			time.Sleep(time.Duration(duration) * time.Millisecond)

			endTime := time.Now().Unix()
			logrus.Info("pile ", p.PileId, " charging car ", car.carId, " finish at ", endTime)
			t := p.startTime()
			logrus.Info(" pile ", p.PileId, " start at ", t)
			if t < startTime && t > 0 {
				logrus.Debug("pile ", p.PileId, " finish car ", car.carId, " from ", startTime, " to ", endTime)
				quantity := float64(car.chargingQuantity)
				p.ChargeTotalCnt++
				p.ChargeTotalQuantity += quantity
				p.CarsLock.Lock()

				//TODO: finish a charing: set the bill finish and other things here
				//TODO: when add codes notice that no blocking alows here
				// 结束充电
				bill, _ := db.GetBillFromBillId(context.Background(), car.carId)

				TimeNow := time.Now().Format(constants.TimeLayoutStr) // end_time
				loc, _ := time.LoadLocation("Local")
				start_time, _ := time.ParseInLocation(constants.TimeLayoutStr, bill.StartTime, loc)
				time_now, _ := time.ParseInLocation(constants.TimeLayoutStr, TimeNow, loc)
				dur := time_now.Sub(start_time).Nanoseconds() * constants.Scale // 实际差了多少ns
				ns, _ := time.ParseDuration("1ns")
				end_time := start_time.Add(ns * time.Duration(dur)) // 实际结束时间
				bill.EndTime = end_time.Format(constants.TimeLayoutStr)

				duration := end_time.Sub(start_time)
				bill.ChargeTime = duration.String() // 充电持续时间

				power := 10
				if bill.ChargeType == constants.QuickCharge {
					power = 30
				}
				bill.ChargeQuantity = duration.Hours() * float64(power) // 充电量

				bill.ServiceFee = 0.8 * bill.ChargeQuantity // 三个费用
				bill.ChargeFee = CalChargeFee(bill.StartTime, bill.EndTime, power)
				bill.TotalFee = bill.ServiceFee + bill.ChargeFee

				err := db.UpdateBill(context.Background(), bill)
				if err != nil {
					logrus.Debug(err)
				}

				p.CarsLock.Unlock()
				p.StartChargeNext()
				p.Signal.Release(1)
			}
		}

	}()

}

func GetPileByTypeTag(pileType int64, pileTag int64) *Pile {
	var p *Pile
	var piles *list.List

	if pileType == constants.ChargingType_Fast {
		piles = S.fastCharingPile
	} else if pileType == constants.ChargingType_Trickle {
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

func (p *Pile) chargeTime(quantity float64) int64 {
	return int64(quantity / float64(p.Power*MUL) * 3600 * 1000)
}

func CalChargeFee(start string, end string, power int) float64 {
	loc, _ := time.LoadLocation("Local")
	// start 是 2022-06-01 15:14:56 格式
	// start_time 是 2022-06-01 15:14:56 +0800 CST
	start_time, _ := time.ParseInLocation(constants.TimeLayoutStr, start, loc) // time.Time格式
	end_time, _ := time.ParseInLocation(constants.TimeLayoutStr, end, loc)
	arr_fee := [7]float64{0.4, 0.7, 1.0, 0.7, 1.0, 0.7, 0.4}
	arr_time := [7]int{7, 10, 15, 18, 21, 23, 24}
	fee := 0.0
	// 所有的时间差都小于24h，但存在跨天的情况
	if start_time.Hour() >= end_time.Hour() { // 跨天 eg.startH=23, endH=16
		// 先算start到当天24点的价格
		next_day := end[0:strings.Index(end, " ")] + " 00:00:00"
		s_index := GetIndex(start_time.Hour()) // eg.16对应arr_fee下标3
		fee1 := CalHelper(start, next_day, s_index, 6, power, arr_fee, arr_time)

		// 再算0点到end的价格
		e_index := GetIndex(end_time.Hour())
		fee2 := CalHelper(next_day, end, 0, e_index, power, arr_fee, arr_time)

		fee = fee1 + fee2
	} else { // 不跨天
		s_index := GetIndex(start_time.Hour())
		e_index := GetIndex(end_time.Hour())
		fee = CalHelper(start, end, s_index, e_index, power, arr_fee, arr_time)
	}
	return fee // 元
}

func GetIndex(hour int) int {
	if hour >= 0 && hour < 7 {
		return 0
	}
	if hour >= 7 && hour < 10 {
		return 1
	}
	if hour >= 10 && hour < 15 {
		return 2
	}
	if hour >= 15 && hour < 18 {
		return 3
	}
	if hour >= 18 && hour < 21 {
		return 4
	}
	if hour >= 21 && hour < 23 {
		return 5
	}
	if hour >= 23 && hour < 24 {
		return 6
	} else {
		return -1
	}
}

func CalHelper(start string, end string, s_index int, e_index int, power int, arr_f [7]float64, arr_t [7]int) float64 {
	loc, _ := time.LoadLocation("Local")
	// eg.start 2022-06-21 16:14:56; end 2022-06-22 7:05:32
	fee := 0.0
	for i := s_index; i <= e_index; i++ {
		if i == s_index {
			if i == e_index {
				s_time, _ := time.ParseInLocation(constants.TimeLayoutStr, start, loc)
				e_time, _ := time.ParseInLocation(constants.TimeLayoutStr, end, loc)
				dur := e_time.Sub(s_time).Hours()
				fee += dur * arr_f[i]
			} else {
				t := arr_t[i]
				mid := start[0:strings.Index(start, " ")] + " " + strconv.Itoa(t) + ":00:00"
				s_time, _ := time.ParseInLocation(constants.TimeLayoutStr, start, loc)
				m_time, _ := time.ParseInLocation(constants.TimeLayoutStr, mid, loc)
				dur := m_time.Sub(s_time).Hours()
				fee += dur * arr_f[i]
			}

		} else if i == e_index {
			t := arr_t[i-1]
			mid := end[0:strings.Index(end, " ")] + " " + strconv.Itoa(t) + ":00:00"
			e_time, _ := time.ParseInLocation(constants.TimeLayoutStr, end, loc)
			m_time, _ := time.ParseInLocation(constants.TimeLayoutStr, mid, loc)
			dur := e_time.Sub(m_time).Hours()
			fee += dur * arr_f[i]
		} else {
			fee += arr_f[i] * (float64(arr_t[i] - arr_t[i-1]))
		}
	}
	return float64(power) * fee
}
