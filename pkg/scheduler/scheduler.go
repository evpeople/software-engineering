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
	DefaultFastPower              = 30
	DefaultTricklePower           = 10
)

var S Scheduler
var nextQueueId int64 = 1 //queue id for the next coming car

func Init() {
	S.trickleChargingPileNum = DefaultTrickleChargingPileNum
	S.fastCharingPileNum = DefaultFastCharingPileNum
	S.waitingAreaSize = DefaultWaitingAreaSize
	S.ChargingQueueLen = DefaultChargingQueueLen
	S.number = 0
	//todo: init by reading config text
	//fastCharingPile
	S.fastCharingPile = list.New()

	for i := 0; i < S.fastCharingPileNum; i++ {
		S.fastCharingPile.PushBack(NewPile(i, S.ChargingQueueLen, ChargingType_Fast, DefaultFastPower, On))
	}
	//trickleChargingPile
	S.trickleChargingPile = list.New()

	for i := 0; i < S.trickleChargingPileNum; i++ {
		S.trickleChargingPile.PushBack(NewPile(i, S.ChargingQueueLen, ChargingType_Trickle, DefaultTricklePower, On))
	}
	S.waitingArea = list.New()
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
	if S.isFull() {
		return 0, 0 //queue if full
	} else {
		S.waitingArea.PushBack(NewCar(userId, carId, nextQueueId, chargingType, chargingQuantity))
		nextQueueId++
		S.number++
		return nextQueueId - 1, S.number
	}

}

//todo: other methods of Scheduler
