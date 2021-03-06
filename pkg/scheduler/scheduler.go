package scheduler

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/evpeople/softEngineer/pkg/constants"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

const (
	WaitingAreaCode               = 0
	ChargingAreaCode              = 1
	DefaultTrickleChargingPileNum = 2
	DefaultFastCharingPileNum     = 3
	DefaultWaitingAreaSize        = 10
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

	S.Lock = sync.Mutex{}
	S.fastPileSignal = semaphore.NewWeighted(int64(S.fastCharingPileNum * S.ChargingQueueLen))
	S.fastWaitingSignal = semaphore.NewWeighted(int64(S.waitingAreaSize))
	S.fastWaitingSignal.Acquire(context.Background(), int64(S.waitingAreaSize))
	S.tricklePileSignal = semaphore.NewWeighted(int64(S.trickleChargingPileNum * S.ChargingQueueLen))
	S.trickleWaitingSignal = semaphore.NewWeighted(int64(S.waitingAreaSize))
	S.trickleWaitingSignal.Acquire(context.Background(), int64(S.waitingAreaSize))
	//todo: init by reading config text

	//fastCharingPile
	S.fastCharingPile = list.New()

	for i := 0; i < S.fastCharingPileNum; i++ {
		S.fastCharingPile.PushBack(NewPile(i, S.ChargingQueueLen, constants.ChargingType_Fast, int64(i+1), DefaultFastPower, On, S.fastPileSignal))
	}
	//trickleChargingPile
	S.trickleChargingPile = list.New()

	for i := 0; i < S.trickleChargingPileNum; i++ {
		S.trickleChargingPile.PushBack(NewPile(i, S.ChargingQueueLen, constants.ChargingType_Trickle, int64(i+1), DefaultTricklePower, On, S.tricklePileSignal))
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

	Lock sync.Mutex //mutex between scheduler threads.
}

//isFull tests if the scheduler can handle more charging request
func (s *Scheduler) isFull() bool {
	return s.WaitingArea.Len() >= s.waitingAreaSize
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
		ChargeType = constants.ChargingType_Fast
		piles = s.fastCharingPile
	} else {
		PileSignal = s.tricklePileSignal
		WaitSignal = s.trickleWaitingSignal
		ChargeType = constants.ChargingType_Trickle
		piles = s.trickleChargingPile
	}

	for {
		logrus.Debug("#waiting", ChargeType)
		PileSignal.Acquire(context.Background(), 1)
		logrus.Debug("#acq pile", ChargeType)
		WaitSignal.Acquire(context.Background(), 1)
		logrus.Debug("#acq wait", ChargeType)
		s.Lock.Lock()
		//firstWaitingCar
		logrus.Info("#go find a car to charging:")
		for e := s.WaitingArea.Front(); e != nil; e = e.Next() {
			logrus.Debug("!!!")
			if car, ok := e.Value.(*Car); ok && car.chargingType == ChargeType {
				pile := chooseAPile(piles)
				pile.CarsLock.Lock()
				pile.WaitingArea.PushBack(car)
				logrus.Info("#car ", car.carId, " is sending to pile(", pile.emptyTimePredict, ") ", pile.PileId)
				if pile.WaitingArea.Len() == 1 && pile.ChargingCar == nil { //pile is empty
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
		s.Lock.Unlock()
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

func GetWaitingInChargeArea(chargeType int) int {
	num := 0
	var piles *list.List

	if chargeType == constants.ChargingType_Fast {
		piles = S.fastCharingPile
	} else {
		piles = S.trickleChargingPile
	}
	for p := piles.Front(); p != nil; p = p.Next() {
		if pile, ok := p.Value.(*Pile); ok {
			num += pile.WaitingArea.Len()
		} else {
			return -1
		}
	}
	return num
}

func GetAllWaiting(chargeType int) int {
	num := 0
	for i := S.WaitingArea.Front(); i != nil; i = i.Next() {
		if car, ok := i.Value.(*Car); ok {
			if car.chargingType == chargeType {
				num++
			}
		} else {
			return -1
		}
	}
	num += GetWaitingInChargeArea(chargeType)
	return num
}

func GetQueueInfoByCarId(carId int) (int, int, int) {
	fastNum := 0
	trickleNum := 0
	var num, queueId, area int

	for i := S.WaitingArea.Front(); i != nil; i = i.Next() {
		if car, ok := i.Value.(*Car); ok {
			if car.carId == int64(carId) {
				if car.chargingType == constants.ChargingType_Fast {
					num = fastNum + GetWaitingInChargeArea(car.chargingType)
				} else {
					num = trickleNum + GetWaitingInChargeArea(car.chargingType)
				}
				queueId = int(car.queueId)
				area = WaitingAreaCode
				break
			}
			if car.chargingType == constants.ChargingType_Fast {
				fastNum++
			} else {
				trickleNum++
			}
		} else {
			return -1, -1, -1
		}
	}
	return queueId, num, area
}

//todo: other methods of Scheduler

//whenCarComing trys to put the car in the queue, if the queue is full return false else return true
func WhenCarComing(userId int64, carId int64, chargingType int, chargingQuantity float64) (int64, int, bool) {
	resp := false
	queueId := int64(-1)
	num := -1

	if chargingQuantity < 0 { //change type
		S.Lock.Lock()
		for e := S.WaitingArea.Front(); e != nil; e = e.Next() {
			if car, ok := e.Value.(*Car); ok && car.carId == carId {
				S.WaitingArea.Remove(e)
				logrus.Debug("changing car type ", carId, "from", car.chargingType, " to ", chargingType)
				car.chargingType = chargingType
				S.WaitingArea.PushBack(car)
				resp = true
				nextQueueId++
				queueId = nextQueueId - 1
				num = S.number
				break
			}
		}
		S.Lock.Unlock()
	} else if chargingType == constants.ChargingType_ChangeQuantity { //change quantity
		S.Lock.Lock()
		for e := S.WaitingArea.Front(); e != nil; e = e.Next() {
			if car, ok := e.Value.(*Car); ok && car.carId == carId {
				logrus.Debug("changing car quantity ", carId, "from", car.carId, " to ", chargingQuantity)
				car.chargingQuantity = chargingQuantity
				resp = true
				queueId = car.queueId
				num = S.number
				break
			}
		}
		S.Lock.Unlock()
	} else if !S.isFull() {
		S.Lock.Lock()
		S.WaitingArea.PushBack(NewCar(userId, carId, nextQueueId, chargingType, chargingQuantity))
		nextQueueId++
		S.number++
		resp = true
		queueId = nextQueueId - 1
		num = S.number
		S.Lock.Unlock()

	}

	if resp {
		if chargingType == constants.ChargingType_Trickle {
			S.trickleWaitingSignal.Release(1)
			logrus.Debug("rel wait", chargingType)
		} else if chargingType == constants.ChargingType_Fast {
			logrus.Debug("rel wait", chargingType)
			S.fastWaitingSignal.Release(1)
		}
	}

	return queueId, num, resp

}

func WhenChargingStop(carId int, pileId int) {
	ans := false
	S.Lock.Lock()
	pile := GetPileById(pileId)
	pile.CarsLock.Lock()
	logrus.Debug(" this is Before stop :", carId, pileId)
	if pile.ChargingCar != nil {
		if int(pile.ChargingCar.carId) == carId {
			logrus.Debug(" this is After stop 1:", carId, pileId)
			pile.ChargingCar = nil
			pile.reStart()
			ans = true
		}
	}
	if !ans {
		for e := pile.WaitingArea.Front(); e != nil; e = e.Next() {
			if car, ok := e.Value.(*Car); ok && int(car.carId) == carId {
				logrus.Debug(" this is After stop 2:", carId, pileId)
				pile.WaitingArea.Remove(e)
				if car.chargingType == constants.ChargingType_Fast {
					S.fastPileSignal.Release(1)
				}
				if car.chargingType == constants.ChargingType_Trickle {
					S.tricklePileSignal.Release(1)
				}
				ans = true
			}
		}
	}
	if !ans {
		for e := S.WaitingArea.Front(); e != nil; e = e.Next() {
			if car, ok := e.Value.(*Car); ok && int(car.carId) == carId {
				logrus.Debug(" this is After stop 3:", carId, pileId)

				S.WaitingArea.Remove(e)
				ans = true
			}

		}
	}

	pile.CarsLock.Unlock()

	S.Lock.Unlock()
}

func ResetPileState(pileId int) {
	pile := GetPileById(pileId)
	S.Lock.Lock()
	if pile != nil {
		if pile.isAlive() {
			pile.shutdown()
			if pile.ChargingCar != nil {
				pile.CarsLock.Lock()
				car := pile.ChargingCar
				WhenFinishCharging(car.carId)

				S.WaitingArea.PushFront(NewCar(car.userId, car.carId, nextQueueId, car.chargingType, car.chargingQuantity))
				nextQueueId++
				S.number++

				pile.CarsLock.Unlock()
			}
		} else {
			pile.reStart()
		}
	}
	S.Lock.Unlock()
}
