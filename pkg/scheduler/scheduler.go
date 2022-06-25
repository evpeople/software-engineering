package scheduler

import (
	"container/list"
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
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
var pileReadyChan chan bool

func Init() {
	S.trickleChargingPileNum = DefaultTrickleChargingPileNum
	S.fastCharingPileNum = DefaultFastCharingPileNum
	S.WaitingAreaSize = DefaultWaitingAreaSize
	S.ChargingQueueLen = DefaultChargingQueueLen
	S.number = 0
	S.lock = sync.Mutex{}
	S.fastPileSignal = semaphore.NewWeighted(int64(S.fastCharingPileNum * S.ChargingQueueLen))
	S.fastWaitingSignal = semaphore.NewWeighted(0)
	S.tricklePileSignal = semaphore.NewWeighted(int64(S.fastCharingPileNum * S.ChargingQueueLen))
	S.trickleWaitingSignal = semaphore.NewWeighted(0)
	//todo: init by reading config text
	//fastCharingPile
	S.fastCharingPile = list.New()

	for i := 0; i < S.fastCharingPileNum; i++ {
		S.fastCharingPile.PushBack(NewPile(i, S.ChargingQueueLen, ChargingType_Fast, DefaultFastPower, On, S.fastPileSignal))
	}
	//trickleChargingPile
	S.trickleChargingPile = list.New()

	for i := 0; i < S.trickleChargingPileNum; i++ {
		S.trickleChargingPile.PushBack(NewPile(i, S.ChargingQueueLen, ChargingType_Trickle, DefaultTricklePower, On, S.tricklePileSignal))
	}
	S.WaitingArea = list.New()

	S.RunScheduler()
}

type Scheduler struct {
	fastPileSignal       *semaphore.Weighted
	fastWaitingSignal    *semaphore.Weighted
	tricklePileSignal    *semaphore.Weighted
	trickleWaitingSignal *semaphore.Weighted

	lock                   sync.Mutex //mutex between scheduler threads.
	number                 int        //the number of the last car entered the waiting area
	trickleChargingPileNum int        //trickle means slow
	fastCharingPileNum     int
	WaitingAreaSize        int
	ChargingQueueLen       int
	trickleChargingPile    *list.List
	fastCharingPile        *list.List

	WaitingArea  *list.List
	eventChannel *chan Event
}

//isFull tests if the scheduler can handle more charging request
func (s *Scheduler) isFull() bool {
	return s.WaitingArea.Len() >= s.WaitingAreaSize
}

//whenCarComing trys to put the car in the queue, if the queue is full return false else return true
func WhenCarComing(userId int64, carId int64, chargingType int, chargingQuantity int) (int64, int) {
	if S.isFull() {
		return 0, 0 //queue if full
	} else {
		S.lock.Lock()
		S.WaitingArea.PushBack(NewCar(userId, carId, nextQueueId, chargingType, chargingQuantity))
		nextQueueId++
		S.number++
		S.lock.Unlock()
		if chargingType == ChargingType_Trickle {
			S.trickleWaitingSignal.Release(1)
		}
		if chargingType == ChargingType_Fast {
			S.fastWaitingSignal.Release(1)
		}
		return nextQueueId - 1, S.number
	}

}

func queryFor(userId int64) {

}

func (s *Scheduler) RunScheduler() {
	s.runFastOrTrickle(true)
	s.runFastOrTrickle(false)

}

func (s *Scheduler) runFastOrTrickle(fast bool) {
	var PileSignal *semaphore.Weighted
	var WaitSignal *semaphore.Weighted
	var ChargeType int
	var piles *list.List
	if fast {
		PileSignal = s.fastPileSignal
		WaitSignal = s.fastWaitingSignal
		ChargeType = ChargingType_Fast
		piles = s.fastCharingPile
	} else {
		PileSignal = s.tricklePileSignal
		WaitSignal = s.trickleWaitingSignal
		ChargeType = ChargingType_Trickle
		piles = s.trickleChargingPile
	}

	for {
		PileSignal.Acquire(context.Background(), 1)
		WaitSignal.Acquire(context.Background(), 1)
		s.lock.Lock()
		//firstWaitingCar
		for e := s.WaitingArea.Front(); e != nil; e = e.Next() {
			if car, ok := e.Value.(Car); ok && car.chargingType == ChargeType {
				pile := chooseAPile(piles)
				pile.CarsLock.Lock()
				pile.WaitingArea.PushBack(car)
				if pile.WaitingArea.Len() == 1 && pile.chargingCar == nil { //pile is empty
					pile.StartChargeNext()
				}
				pile.CarsLock.Unlock()
				break
			}
		}
		s.lock.Unlock()
	}

}

func chooseAPile(piles *list.List) *Pile {
	var finishTimePredict int64
	finishTimePredict = 0
	var bestPile *Pile
	bestPile = nil
	for e := piles.Front(); e != nil; e = e.Next() {
		if p, ok := e.Value.(Pile); ok {
			if bestPile == nil {
				p.CarsLock.Lock()
				if p.WaitingArea.Len() >= p.MaxWaitingNum {
					p.CarsLock.Unlock()
					continue
				}
				finishTimePredict = p.emptyTimePredict
				p.CarsLock.Unlock()
				bestPile = &p
			} else {
				p.CarsLock.Lock()
				if p.WaitingArea.Len() >= p.MaxWaitingNum || p.emptyTimePredict > finishTimePredict {
					p.CarsLock.Unlock()
					continue
				}
				finishTimePredict = p.emptyTimePredict
				p.CarsLock.Unlock()
				bestPile = &p
			}
		}
	}
	return bestPile
}

/*
//shit code never use :ru
func (s *Scheduler) run() {
	for {
		if s.WaitingArea.Front() == nil { //if no waiting car
			logrus.Debug("there is no car in the WaitingArea.")
			s.mutex.Lock() //P
		}
		nextCar := s.WaitingArea.Front().Value.(Car)
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
			s.WaitingArea.Remove(s.WaitingArea.Front())
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
