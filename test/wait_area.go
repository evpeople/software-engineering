package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type waitAreaQuest struct {
	CarId    int `json:"car_id"`
	Ctype    int `json:"ctype"`
	Quantity int `json:"quantity"`
}

func GetWaitArea() {
	// URL := "http://122.9.146.200:8080/v1"
	resp, err := http.Get(URL + "/charge/list")
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
