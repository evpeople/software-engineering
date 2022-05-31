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

//todo: other methods of Car
