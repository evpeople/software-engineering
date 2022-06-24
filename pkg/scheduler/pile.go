package scheduler

import (
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
	Channel             chan Car
	Signals 					Signals
	// 充电时长（小时）=实际充电度数/充电功率(度/小时)，需要的时候再计算
}



func NewPile(pileId int, maxWaitingNum int, pileType int, power float32, status PileStatus,askForPileReady*chan bool) *Pile {
	return &Pile{pileId, maxWaitingNum, pileType, power, status, 0, 0,
		 make(chan Car, maxWaitingNum),Signals{askForPileReady,make(chan bool)}}
	
}

func (p *Pile) run() {
	go func() {
		var car Car
		for {
			select {
			case car = <-p.Channel:
			default:
				*p.Signals.isPileReady<-true
				logrus.Debug("pile ",p.PileId," is empty")
				car =  <-p.Channel
			}
			logrus.Info("pile ",p.PileId," got car ",car.carId,"Start ")
			t:=float32(car.chargingQuantity)/p.Power
			time.Sleep(time.Duration(t)*time.Second)
			select {
			case <-p.Signals.stopPile:
				logrus.Debug("pile ",p.PileId," is stoped")
				break
			default:
				p.ChargeTotalCnt++
				p.ChargeTotalQuantity+=float64(car.chargingQuantity)
				//TODO: finish a charing: set the bill finish here
			}
		}
	}()
}

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
