package scheduler

import (
	"container/list"
)

const (
	ChargingType_Fast    = 0
	ChargingType_Trickle = 1

	DefaultTrickleChargingPileNum = 2
	DefaultFastCharingPileNum     = 3
	DefaultWaitingAreaSize        = 100
	DefaultChargingQueueLen       = 3
	DefaultFastPower              = 30
	DefaultTricklePower           = 10
)

var s Scheduler
var nextQueueId int64 = 1 //queue id for the next coming car

func Init() {
	s.trickleChargingPileNum = DefaultTrickleChargingPileNum
	s.fastCharingPileNum = DefaultFastCharingPileNum
	s.waitingAreaSize = DefaultWaitingAreaSize
	s.ChargingQueueLen = DefaultChargingQueueLen
	s.number = 0
	//todo: init by reading config text
	//fastCharingPile
	s.fastCharingPile = list.New()

	for i := 0; i < s.fastCharingPileNum; i++ {
		s.fastCharingPile.PushBack(NewPile(i, s.ChargingQueueLen, ChargingType_Fast, int64(i+1), DefaultFastPower, On))
	}
	//trickleChargingPile
	s.trickleChargingPile = list.New()

	for i := 0; i < s.trickleChargingPileNum; i++ {
		s.trickleChargingPile.PushBack(NewPile(i, s.ChargingQueueLen, ChargingType_Trickle, int64(i+1), DefaultTricklePower, On))
	}
	s.waitingArea = list.New()
}

type Scheduler struct {
	number                 int //the number of the last car entered the waiting area
	trickleChargingPileNum int //trickle means slow
	fastCharingPileNum     int
	waitingAreaSize        int
	ChargingQueueLen       int
	trickleChargingPile    *list.List
	fastCharingPile        *list.List
	waitingArea            *list.List
}

//isFull tests if the scheduler can handle more charging request
func (s *Scheduler) isFull() bool {
	return s.waitingArea.Len() >= s.waitingAreaSize
}

//whenCarComing trys to put the car in the queue, if the queue is full return false else return true
func WhenCarComing(userId int64, carId int64, chargingType int, chargingQuantity int) (int64, int) {
	if s.isFull() {
		return 0, 0 //queue if full
	} else {
		s.waitingArea.PushBack(NewCar(userId, carId, nextQueueId, chargingType, chargingQuantity))
		nextQueueId++
		s.number++
		return nextQueueId - 1, s.number
	}

}

//todo: other methods of Scheduler
