package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Event struct {
	Type  string `json:"Type"`
	Id    string `json:"id"`
	CType string `json:"CType"`
	Num   int    `json:"Num"`
}

var URL string

func main() {
	// 打开json文件
	URL = "http://122.9.146.200:8080/v1/"
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
				charge_quantity := v.Num
				carID := getCarID(v.Id)
				if charge_quantity == 0 {
					stopCharge(carID)
				}
				carIdInt, _ := strconv.Atoi(carID)
				charge_Type := getChargeType(v.CType)
				sendCharge(carIdInt, charge_Type, charge_quantity)
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
func getCarID(a string) string {
	a = a[1:]
	return a
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

func getChargeType(a string) (ctype int) {
	switch a[0] {
	case 'F':
		ctype = 0
	case 'T':
		ctype = 1
	case 'O':
		ctype = 2
	}
	return
}
func stopCharge(carID string) {
	data := make(map[string]interface{})
	data["car_id"] = carID
	bytesData, _ := json.Marshal(data)
	resp, _ := http.Post(URL+"/charge/stop", "application/json", bytes.NewReader(bytesData))
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
func sendCharge(id, typ, quantity int) {
	data := make(map[string]interface{})
	data["car_id"] = id
	data["charging_type"] = typ
	data["charging_quantity"] = quantity
	bytesData, _ := json.Marshal(data)
	resp, _ := http.Post(URL+"/charge/come", "application/json", bytes.NewReader(bytesData))
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
