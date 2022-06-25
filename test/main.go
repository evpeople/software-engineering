package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
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
type Pile struct {
	Pile_type int
	Pile_tag  int
}

var URL string
var Token string
var PilesWrite *bufio.Writer
var WaitWrite *bufio.Writer

func main() {
	// 打开json文件
	// URL = "http://122.9.146.200:8080/v1"
	URL = "http://192.168.147.122:8080/v1"
	Token = "?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MTksImV4cCI6MTY1NjE0NDI1OSwib3JpZ19pYXQiOjE2NTYxNDA2NTl9.EAGDoG5beb1hblLD6MiQmPoAoUkM2VBUdFHdMhqdtew"
	jsonFile, _ := os.Open("data.json")

	PilesFilePath := "./piles.txt"
	PilesFile, err := os.OpenFile(PilesFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer PilesFile.Close()
	//写入文件时，使用带缓存的 *Writer
	PilesWrite = bufio.NewWriter(PilesFile)
	WaitingFilePath := "./waiting.txt"
	WaitFile, err := os.OpenFile(WaitingFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer WaitFile.Close()
	//写入文件时，使用带缓存的 *Writer
	WaitWrite = bufio.NewWriter(WaitFile)
	//Flush将缓存的文件真正写入到文件中

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
		// fmt.Println(v.Type)
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
		go getWaitArea()
		go getWaitChargeCar()
		time.Sleep(30 * time.Second)
	}
	PilesWrite.Flush()
	WaitWrite.Flush()
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
	resp, err := http.Get(URL + "/charge/list" + Token)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// var out bytes.Buffer
	// json.Indent(&out, body, "", "\t")
	// fmt.Printf("student=%v\n", out.String())
	// fmt.Println(string(body))
	// WaitWrite.WriteString(out.String())
	WaitWrite.Write(body)
	WaitWrite.WriteString("\n\n\n")
	WaitWrite.Flush()
	var res []WaitAreaQuest
	_ = json.Unmarshal(body, &res)
	fmt.Println(res)

}
func getWaitChargeCar() {
	piles := [5]Pile{

		{
			Pile_type: 0,
			Pile_tag:  1,
		},
		{
			Pile_type: 0,
			Pile_tag:  2,
		},
		{
			Pile_type: 0,
			Pile_tag:  3,
		},
		{
			Pile_type: 1,
			Pile_tag:  1,
		},
		{
			Pile_type: 1,
			Pile_tag:  2,
		},
	}

	for _, v := range piles {
		u := fmt.Sprintf(URL+"/admin/cars"+Token+"&pile_type=%d&pile_tag=%d", v.Pile_type, v.Pile_tag)
		resp, err := http.Get(u)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		// var out bytes.Buffer
		// json.Indent(&out, body, "", "\t")
		// fmt.Printf("student=%v\n", out.String())
		PilesWrite.Write(body)
		PilesWrite.WriteByte('\n')
		// fmt.Println(string(body))
		var res []WaitAreaQuest
		_ = json.Unmarshal(body, &res)
		// fmt.Println(res)
	}
	PilesWrite.Flush()
}
