package scheduler

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
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

var S Scheduler
var nextQueueId int64 = 1 //queue id for the next coming car

func Init() {
	S.trickleChargingPileNum = DefaultTrickleChargingPileNum
	S.fastCharingPileNum = DefaultFastCharingPileNum
	S.waitingAreaSize = DefaultWaitingAreaSize
	S.ChargingQueueLen = DefaultChargingQueueLen
	S.number = 0

	S.lock = sync.Mutex{}
	S.fastPileSignal = semaphore.NewWeighted(int64(S.fastCharingPileNum * S.ChargingQueueLen))
	S.fastWaitingSignal = semaphore.NewWeighted(int64(S.waitingAreaSize))
	S.fastWaitingSignal.Acquire(context.Background(), int64(S.waitingAreaSize))
	S.tricklePileSignal = semaphore.NewWeighted(int64(S.fastCharingPileNum * S.ChargingQueueLen))
	S.trickleWaitingSignal = semaphore.NewWeighted(int64(S.waitingAreaSize))
	S.trickleWaitingSignal.Acquire(context.Background(), int64(S.waitingAreaSize))
	//todo: init by reading config text

	//fastCharingPile
	S.fastCharingPile = list.New()

	for i := 0; i < S.fastCharingPileNum; i++ {
		S.fastCharingPile.PushBack(NewPile(i, S.ChargingQueueLen, ChargingType_Fast, int64(i+1), DefaultFastPower, On, S.fastPileSignal))
	}
	//trickleChargingPile
	S.trickleChargingPile = list.New()

	for i := 0; i < S.trickleChargingPileNum; i++ {
		S.trickleChargingPile.PushBack(NewPile(i, S.ChargingQueueLen, ChargingType_Trickle, int64(i+1), DefaultTricklePower, On, S.tricklePileSignal))
	}
	S.WaitingArea = list.New()
	S.RunScheduler()
	logrus.Info("init scheduler over.")
}

type Scheduler struct {
	number                 int //the number of the last car entered the waiting area
	trickleChargingPileNum int //trickle means slow
	fastCharingPileNum     int
	waitingAreaSize        int
	ChargingQueueLen       int
	trickleChargingPile    *list.List
	fastCharingPile        *list.List
	WaitingArea            *list.List

	fastPileSignal       *semaphore.Weighted
	fastWaitingSignal    *semaphore.Weighted
	tricklePileSignal    *semaphore.Weighted
	trickleWaitingSignal *semaphore.Weighted

	lock sync.Mutex //mutex between scheduler threads.
}

//isFull tests if the scheduler can handle more charging request
func (s *Scheduler) isFull() bool {
	return s.WaitingArea.Len() >= s.waitingAreaSize
}

//whenCarComing trys to put the car in the queue, if the queue is full return false else return true
func WhenCarComing(userId int64, carId int64, chargingType int, chargingQuantity float64) (int64, int) {
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
			logrus.Debug("rel wait", chargingType)
		}
		if chargingType == ChargingType_Fast {
			logrus.Debug("rel wait", chargingType)
			S.fastWaitingSignal.Release(1)
		}
		return nextQueueId - 1, S.number
	}

}

func queryFor(userId int64) {

}

func (s *Scheduler) RunScheduler() {
	go func() {
		s.runFastOrTrickle(true)
	}()
	go func() {
		s.runFastOrTrickle(false)
	}()

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
		logrus.Debug("#waiting", ChargeType)
		PileSignal.Acquire(context.Background(), 1)
		logrus.Debug("#acq pile", ChargeType)
		WaitSignal.Acquire(context.Background(), 1)
		logrus.Debug("#acq wait", ChargeType)
		s.lock.Lock()
		//firstWaitingCar
		logrus.Info("#go find a car to charging:")
		for e := s.WaitingArea.Front(); e != nil; e = e.Next() {
			logrus.Debug("!!!")
			if car, ok := e.Value.(*Car); ok && car.chargingType == ChargeType {
				pile := chooseAPile(piles)
				pile.CarsLock.Lock()
				pile.WaitingArea.PushBack(car)
				logrus.Info("#car ", car.carId, " is sending to pile(", pile.emptyTimePredict, ") ", pile.PileId)
				if pile.WaitingArea.Len() == 1 && pile.chargingCar == nil { //pile is empty
					logrus.Debug("pile ", pile.PileId, " has a queue now. ")
					pile.StartChargeNext()
					pile.emptyTimePredict = time.Now().Unix() + pile.chargeTime(car.chargingQuantity)
				} else {
					pile.emptyTimePredict += (int64(float32(car.chargingQuantity) / pile.Power))
				}
				s.WaitingArea.Remove(e)
				//TODO:car enter pile place here
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
		if p, ok := e.Value.(*Pile); ok {
			if bestPile == nil {
				p.CarsLock.Lock()
				if p.WaitingArea.Len() >= p.MaxWaitingNum {
					p.CarsLock.Unlock()
					continue
				}
				finishTimePredict = p.emptyTimePredict
				p.CarsLock.Unlock()
				bestPile = p
			} else {
				p.CarsLock.Lock()
				if p.WaitingArea.Len() >= p.MaxWaitingNum || p.emptyTimePredict >= finishTimePredict {
					p.CarsLock.Unlock()
					continue
				}
				finishTimePredict = p.emptyTimePredict
				p.CarsLock.Unlock()
				bestPile = p
			}
		}
	}
	return bestPile
}

//todo: other methods of Scheduler
