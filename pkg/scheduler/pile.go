package scheduler

import (
	"container/list"
	// "github.com/evpeople/softEngineer/pkg/errno"
)

// 充电桩状态的枚举类型
type PileStatus int

const (
	On        PileStatus = iota + 1 // EnumIndex = 1，充电桩开启
	Off                             // EnumIndex = 2，充电桩关闭
	Charging                        // EnumIndex = 3，充电桩正在充电
	Breakdown                       // EnumIndex = 4，充电桩故障
)

// 返回充电桩状态对应的字符串
func (status PileStatus) String() string {
	return [...]string{"On", "Off", "Charging", "Breakdown"}[status-1]
}

// 返回充电桩状态对应索引值
func (status PileStatus) EnumIndex() int {
	return int(status)
}

type Pile struct {
	PileId              int
	MaxWaitingNum       int
	Type                int64
	PileTag             int64
	Power               float32
	Status              PileStatus
	ChargeTotalCnt      int
	ChargeTotalQuantity float64
	ChargeArea          *list.List

	// 充电时长（小时）=实际充电度数/充电功率(度/小时)，需要的时候再计算
}

// // 判断当前充电桩的队列是否满
// func (p *Pile) isFull() bool {
// 	return p.ChargeArea.Len() >= p.MaxWaitingNum
// }

// func (p *Pile) close() (bool, errno.ErrNo) {
// 	switch p.Status {
// 	case Off:
// 		return true, errno.Success
// 	case On:
// 		p.Status = Off
// 		return true, errno.Success
// 	case Breakdown:
// 		return false, errno.TurnOffBreakdownPileErr
// 	case Charging:
// 		// ? 需要考虑，充电中能否强制关机？能的话，需要添加后续处理；不能的话，需要返回错误信息
// 		// 此处暂时作为 充电中不能关机处理
// 		return false, errno.TurnOffChargingPileErr
// 	default:
// 		return true, errno.Success // 默认 Status 字段未初始化时，充电桩处于关闭状态
// 	}
// }

// func (p *Pile) open() (bool, errno.ErrNo) {
// 	switch p.Status {
// 	case Off:
// 		p.Status = On
// 		return true, errno.Success
// 	case On:
// 		return true, errno.Success
// 	case Breakdown:
// 		return false, errno.TurnOffBreakdownPileErr
// 	case Charging:
// 		// ? 需要考虑，充电中能否强制关机？能的话，需要添加后续处理；不能的话，需要返回错误信息
// 		// 此处暂时作为 充电中不能关机处理
// 		return false, errno.TurnOffChargingPileErr
// 	default:
// 		return true, errno.Success // 默认 Status 字段未初始化时，充电桩处于关闭状态
// 	}
// }

func NewPile(pileId int, maxWaitingNum int, pileType int64, pileTag int64, power float32, status PileStatus) *Pile {
	return &Pile{pileId, maxWaitingNum, pileType, pileTag, power, status, 0, 0, list.New()}
}

func GetPileByTypeTag(pileType int64, pileTag int64) *Pile {
	var p *Pile
	var piles *list.List

	if pileType == ChargingType_Fast {
		piles = S.fastCharingPile
	} else if pileType == ChargingType_Trickle {
		piles = S.trickleChargingPile
	} else {
		return nil
	}

	for i := piles.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileTag == pileTag {
			return p
		}
	}
	return nil
}

func GetPileById(pileId int) *Pile {
	var p *Pile
	// 遍历慢充桩
	for i := S.trickleChargingPile.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileId == pileId {
			return p
		}
	}
	// 遍历快充桩
	for i := S.fastCharingPile.Front(); i != nil; i = i.Next() {
		p = i.Value.(*Pile)
		if p.PileId == pileId {
			return p
		}
	}
	return nil
}
