package scheduler

const (
	chargeFinish = "chargeFinish"
)

type Event struct {
	carId     int64
	pileId    int
	startTime int64
	endTime   int64
	quantity  float64
}

func NewChargeFinishEvent(carId int64, pileId int, start_time int64, end_time int64, power float32) *Event {
	return &Event{carId, pileId, start_time, end_time, float64(power) * float64((end_time-start_time)/1)}
}
