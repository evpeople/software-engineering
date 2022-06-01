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
	s.number = 0
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
}

type Scheduler struct {
	number                 int //the number of the last car entered the waiting area
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
func WhenCarComing(userId int64, carId int64, chargingType int, chargingQuantity int) int {
	if s.isFull() {
		return -1 //queue if full
	} else {
		s.waitingArea.PushBack(NewCar(userId, carId, chargingType, chargingQuantity))
		s.number++
		return s.number
	}

}

//todo: other methods of Scheduler
