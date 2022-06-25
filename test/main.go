package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
type WaitAreaQuest struct {
	CarId    int `json:"car_id"`
	Ctype    int `json:"ctype"`
	Quantity int `json:"quantity"`
}

var URL string
var Token string

func main() {
	// 打开json文件
	// URL = "http://122.9.146.200:8080/v1"
	URL = "http://192.168.147.122:8080/v1"
	Token = "?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MTgsImV4cCI6MTY1NjE0MDQzNywib3JpZ19pYXQiOjE2NTYxMzY4Mzd9.xtCJAU6R2ceEGZPj2h5I583KANpiDRCuiQWnot5df08"
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
		fmt.Println(v.Type)
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
				sendPileReset(pileTag, pileType)
			}
		case "C":
			{
				carIdInt, _ := strconv.Atoi(getCarID(v.Id))
				chargeQuantity := v.Num
				chargeType := getChargeType(v.CType)
				sendCharge(carIdInt, chargeType, chargeQuantity)
			}
		}
	}
	getWaitArea()
}

func getCarID(a string) string {
	a = a[1:]
	return a
}

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

func getPileTagTy(a string) (id string, pile_type string) {
	id = a[1:]
	if tag := a[0]; tag == 'F' {
		pile_type = "0"
	} else if tag == 'T' {
		pile_type = "1"
	}
	return
}

func stopCharge(carID string) {
	data := make(map[string]interface{})
	fmt.Println("dsds")
	data["car_id"] = carID
	bytesData, _ := json.Marshal(data)
	resp, _ := http.Post(URL+"/charge/stop"+Token, "application/json", bytes.NewReader(bytesData))
	fmt.Println("aaaa")
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("bbbb")

	fmt.Println(string(body))
}

func sendCharge(id, typ, quantity int) {
	data := make(map[string]interface{})
	data["car_id"] = id
	data["charging_type"] = typ
	data["charging_quantity"] = quantity
	bytesData, _ := json.Marshal(data)
	resp, _ := http.Post(URL+"/charge/come"+Token, "application/json", bytes.NewReader(bytesData))
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func sendPileReset(id string, pile_type string) {
	resp, _ := http.Post(URL+"/admin/pile/"+id+Token+"&pile_type="+pile_type, "application/json", bytes.NewReader([]byte{}))
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
func getWaitArea() {
	// URL := "http://122.9.146.200:8080/v1"
	resp, err := http.Get(URL + "/charge/list" + Token)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	var res []WaitAreaQuest
	_ = json.Unmarshal(body, &res)
	// arrWaitArea := [3]int{res.CarId, res.Ctype, res.Quantity}
	fmt.Println(res)
}
