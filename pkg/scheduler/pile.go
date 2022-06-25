package scheduler

import (
	"container/list"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
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

	Signal *semaphore.Weighted

	PileStartTime     int64
	PileStartTimeLock sync.Mutex

	emptyTimePredict int64
	WaitingArea      *list.List
	chargingCar      *Car
	CarsLock         sync.Mutex //lock for above 3
	// 充电时长（小时）=实际充电度数/充电功率(度/小时)，需要的时候再计算
}

func NewPile(pileId int, maxWaitingNum int, pileType int, power float32, status PileStatus, siganl *semaphore.Weighted) *Pile {
	return &Pile{pileId, maxWaitingNum, pileType, power, status, 0, 0,
		siganl, 0, sync.Mutex{}, time.Now().Unix(), list.New(), nil, sync.Mutex{}}

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
		car, ok := next.Value.(Car)
		if ok {
			p.chargingCar = &car //charging car
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
