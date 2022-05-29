package scheduler

type Car struct {
	userId           string
	carId            string
	chargingType     int
	chargingQuantity int
	isCharging       bool
}

func NewCar(userId string, carId string, chargingType int, chargingQuantity int) *Car {
	return &Car{userId, carId, chargingType, chargingQuantity, false}
}

//todo: other methods of Car
