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
		case "C":
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
