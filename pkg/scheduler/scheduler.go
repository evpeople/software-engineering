package scheduler

import (
	"container/list"
)

const (
	ChargingType_Fast    = 1
	ChargingType_Trickle = 2

	DefaultTrickleChargingPileNum = 2
	DefaultFastCharingPileNum     = 3
	DefaultWaitingAreaSize        = 100
	DefaultChargingQueueLen       = 3
)

var s Scheduler

func Init() {
	s.trickleChargingPileNum = DefaultTrickleChargingPileNum
	s.fastCharingPileNum = DefaultFastCharingPileNum
	s.waitingAreaSize = DefaultWaitingAreaSize
	s.chargingQueueLen = DefaultChargingQueueLen
	//todo: init by reading config text
	//fastCharingPile
	s.fastCharingPile = make([]*list.List, s.fastCharingPileNum)

	for i := 0; i < s.fastCharingPileNum; i++ {
		s.fastCharingPile[i] = list.New()
	}
	//trickleChargingPile
	s.trickleChargingPile = make([]*list.List, s.trickleChargingPileNum)

	for i := 0; i < s.trickleChargingPileNum; i++ {
		s.trickleChargingPile[i] = list.New()
	}
	s.waitingArea = list.New()
	s.waitingArea.PushBack(NewCar("", "", 1, 1))
}

type Scheduler struct {
	trickleChargingPileNum int //trickle means slow
	fastCharingPileNum     int
	waitingAreaSize        int
	chargingQueueLen       int
	trickleChargingPile    []*list.List
	fastCharingPile        []*list.List
	waitingArea            *list.List
}

//isFull tests if the scheduler can handle more charging request
func (s *Scheduler) isFull() bool {
	return s.waitingArea.Len() >= s.waitingAreaSize
}

//whenCarComing trys to put the car in the queue, if the queue is full return false else return true
func WhenCarComing(userId string, carId string, chargingType int, chargingQuantity int) bool {
	if s.isFull() {
		return false
	} else {
		s.waitingArea.PushBack(NewCar(userId, carId, chargingType, chargingQuantity))
		return true
	}

}

//todo: other methods of Scheduler
