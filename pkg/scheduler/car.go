package scheduler

type Car struct {
	userId           int64
	carId            int64
	chargingType     int
	chargingQuantity int
	isCharging       bool
}

func NewCar(userId int64, carId int64, chargingType int, chargingQuantity int) *Car {
	return &Car{userId, carId, chargingType, chargingQuantity, false}

}

func (car *Car) GetUserId() int64 {
	return car.userId
}

func (car *Car) GetCarId() int64 {
	return car.carId
}

func (car *Car) GetChargingQuantity() int {
	return car.chargingQuantity
}

//todo: other methods of Car
