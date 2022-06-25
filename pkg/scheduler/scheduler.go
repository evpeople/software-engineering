package scheduler

import (
	"container/list"
	"sync"

	"github.com/sirupsen/logrus"
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

var s Scheduler
var nextQueueId int64 = 1 //queue id for the next coming car
var pileReadyChan chan bool

func Init() {
	s.trickleChargingPileNum = DefaultTrickleChargingPileNum
	s.fastCharingPileNum = DefaultFastCharingPileNum
	s.waitingAreaSize = DefaultWaitingAreaSize
	s.ChargingQueueLen = DefaultChargingQueueLen
	s.number = 0
	s.mutex = sync.Mutex{}
	//todo: init by reading config text
	//fastCharingPile
	s.fastCharingPile = list.New()

	for i := 0; i < s.fastCharingPileNum; i++ {
		s.fastCharingPile.PushBack(NewPile(i, s.ChargingQueueLen, ChargingType_Fast, DefaultFastPower, On, s.eventChannel))
	}
	//trickleChargingPile
	s.trickleChargingPile = list.New()

	for i := 0; i < s.trickleChargingPileNum; i++ {
		s.trickleChargingPile.PushBack(NewPile(i, s.ChargingQueueLen, ChargingType_Trickle, DefaultTricklePower, On, s.eventChannel))
	}
	s.waitingArea = list.New()
}

type Scheduler struct {
	mutex                  sync.Mutex //mutex between scheduler threads.
	number                 int //the number of the last car entered the waiting area
	trickleChargingPileNum int //trickle means slow
	fastCharingPileNum     int
	waitingAreaSize        int
	ChargingQueueLen       int
	trickleChargingPile    *list.List
	fastCharingPile        *list.List

	waitingArea            *list.List
	eventChannel           *chan Event
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
		s.mutex.Unlock() //V
		return nextQueueId - 1, s.number
	}

}

func queryFor(userId int64) {

}

func (s *Scheduler) runEventsListener() {
	go func() {
		for {
			c := <-*s.eventChannel
			pileId:=c.pileId
			p:=GetPileById(pileId)
			p.ChargeTotalCnt++
			p.ChargeTotalQuantity += float64(car.chargingQuantity)
			//TODO: finish a charing: set the bill finish here
		}
	}()
}

func (s *Scheduler)runScheduler(){

}


/*
//shit code never use :run
func (s *Scheduler) run() {
	for {
		if s.waitingArea.Front() == nil { //if no waiting car
			logrus.Debug("there is no car in the waitingArea.")
			s.mutex.Lock() //P
		}
		nextCar := s.waitingArea.Front().Value.(Car)
		l := s.fastCharingPile
		if nextCar.chargingType == ChargingType_Trickle {
			l = s.trickleChargingPile
		}
		addInPile := false
		for e := l.Front(); e != nil; e = e.Next() {
			if pile, ok := e.Value.(Pile); ok {
				select {
				case pile.Channel <- nextCar:
					logrus.Info("car ", nextCar.carId, " is sending to ")
					addInPile = true
					break
				default:
					logrus.Info("car ", nextCar.carId, " skipping pile", pile.PileId, " because it is full")
				}
			}
		}

		if addInPile {
			s.waitingArea.Remove(s.waitingArea.Front())
		} else {
			<-*s.eventChannel
		}
	}
}
*/
// func (s *Scheduler) stopPile(pileId int) { // stop a pile when it is shut down by admin or of force majeure
// 	pile := GetPileById(pileId)
// 	pile.Signals.stopPile <- true
// }

//todo: other methods of Scheduler
