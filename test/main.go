package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Event struct {
	Type  string `json:"Type"`
	Id    string `json:"id"`
	CType string `json:"CType"`
	Num   int    `json:"Num"`
}

func main() {
	// 打开json文件
	jsonFile, err := os.Open("data.json")

	// 最好要处理以下错误
	if err != nil {
		fmt.Println(err)
	}

	// 要记得关闭
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var event []Event
	json.Unmarshal([]byte(byteValue), &event)

	fmt.Println(event)
	for _, v := range event {
		switch v.Type {
		case "A":
			{
				carID := getCarID(v.Id)
				charge_quantity = v.Num
				charge_Type = getChargeType()
			}
		case "B":
			{
				//默认所有充电桩都是开启状态，忽略Num部分
				pileTag, pileType := getPileTagTy(v.Id)
				postPileStatus(pileTag, pileType)
			}
		case "C":
			{
				carID := getCarID(v.Id)
				chargeQuantity = v.Num
				chargeType = getChargeType()
			}
		}
	}
}
func getCarID(a string) (id int) {
	a = a[1:]
	id, err := strconv.Atoi(a)
	if err != nil {
		panic(err)
	}
	return
}

func getPileTagTy(a string) (id string, pile_type int) {
	id = a[1:]
	if tag := a[0]; tag == 'F' {
		pile_type = 0
	} else if tag == 'T' {
		pile_type = 1
	}
	return
}

func postPileStatus(id string, pile_type int) {

}
