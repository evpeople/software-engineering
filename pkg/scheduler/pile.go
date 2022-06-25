package scheduler

import (
	"container/list"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
	Type                int
	Power               float32
	Status              PileStatus
	ChargeTotalCnt      int
	ChargeTotalQuantity float64

	Channel *chan Event

	EmptyTimePredict int64

	PileStartTime     int64
	PileStartTimeLock sync.Mutex
	cars              *list.List
	// 充电时长（小时）=实际充电度数/充电功率(度/小时)，需要的时候再计算
}

func NewPile(pileId int, maxWaitingNum int, pileType int, power float32, status PileStatus, eventChan *chan Event) *Pile {
	return &Pile{pileId, maxWaitingNum, pileType, power, status, 0, 0,
		eventChan, time.Now().Unix(), 0, sync.Mutex{}, list.New()}

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

func (p *Pile) startCharge(car *Car) {
	go func() {
		logrus.Info("pile ", p.PileId, " got car ", car.carId, "Start ")
		duration := float32(car.chargingQuantity) / p.Power
		startTime := time.Now().Unix()
		time.Sleep(time.Duration(duration) * time.Second)
		endTime := time.Now().Unix()

		t := p.startTime()
		if t < startTime && t > 0 {
			*p.Channel <- *NewChargeFinishEvent(car.carId, p.PileId, startTime, endTime)
		}
	}()
	// p.ChargeTotalCnt++
	// p.ChargeTotalQuantity += float64(car.chargingQuantity)
	// //notTODO: finish a charing: set the bill finish here

}

/*
//shit code never use :run
func (p *Pile) run() {
	go func() {
		var car Car
		for {
			select { // get car from channel
			case car = <-p.Channel:
			default:
				*p.Signals.isPileReady <- true
				logrus.Debug("pile ", p.PileId, " is empty")
				car = <-p.Channel // blocking here
				p.EmptyTimePredict = time.Now().Unix()
			}

			logrus.Info("pile ", p.PileId, " got car ", car.carId, "Start ")
			t := float32(car.chargingQuantity) / p.Power
			time.Sleep(time.Duration(t) * time.Second)

			select {
			case <-p.Signals.stopPile:
				logrus.Debug("pile ", p.PileId, " is stoped")
				break
			default:
				p.ChargeTotalCnt++
				p.ChargeTotalQuantity += float64(car.chargingQuantity)
				//TODO: finish a charing: set the bill finish here
			}
		}
	}()
}
*/

func GetPileById(pileId int) *Pile {
	var p *Pile
	// 遍历慢充桩
	for i := s.trickleChargingPile.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileId == pileId {
			return p
		}
	}
	// 遍历快充桩
	for i := s.fastCharingPile.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileId == pileId {
			return p
		}
	}
	return nil
}
