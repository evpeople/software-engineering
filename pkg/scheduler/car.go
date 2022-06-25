package scheduler

type Car struct {
	userId           int64
	carId            int64
	queueId          int64
	chargingType     int
	chargingQuantity float64
	isCharging       bool
}

func NewCar(userId int64, carId int64, queueId int64, chargingType int, chargingQuantity float64) *Car {
	return &Car{userId, carId, queueId, chargingType, chargingQuantity, false}

}

func (car *Car) GetUserId() int64 {
	return car.userId
}

func (car *Car) GetCarId() int64 {
	return car.carId
}

func (car *Car) GetChargingQuantity() float64 {
	return car.chargingQuantity
}

func (car *Car) GetCType() int {
	return car.chargingType
}

//todo: other methods of Car
