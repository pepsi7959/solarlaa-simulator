package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

/*
"v": 396,
"i": 96.05,
"p": 56.00,
"phaseshift": 0.0,
"status": 1,
"timestamp": "2018-01-01 00:15:00.000"
}
*/

type Consumption struct {
	Timestamp  string  `json:"timestamp"`
	Power      float64 `json:"p"`
	Voltage    float64 `json:"v"`
	Current    float64 `json:"i"`
	Phaseshift float64 `json:"phaseshift"`
	Status     int     `json:"status"`
}

var OWNER_ID string = "2052"
var DEVICE_ID string = "111111111111"
var URL string = "http://solaraa.advancedlogic.co/api/v1/plugs/" + DEVICE_ID + "/owners/2052/consumption"
var RETRY_ON_FAIL int = 5

func sendToServer(c *Consumption) bool {
	fmt.Println("send: ", c)
	tr := &http.Transport{
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	fmt.Println(c.Timestamp)
	jstr, err := json.Marshal(*c)

	if err != nil {
		fmt.Println(err)
		return false
	}

	resp, err := client.Post(URL, "application/json", bytes.NewBuffer(jstr))

	if err != nil {
		fmt.Println(err)
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("status: %d , %s", resp.StatusCode, body)
		return false
	}

	fmt.Printf("recv: %d\n", len(body))
	return true
}

func main() {

	if len(os.Args) != 5 {
		fmt.Println("solarlaa-simulator <ownerid> <deviceid> <csv file> <line number>")
		return
	}

	OWNER_ID = os.Args[1]
	DEVICE_ID = os.Args[2]
	f := os.Args[3]
	l_num, _ := strconv.Atoi(os.Args[4])

	URL = "http://solaraa.advancedlogic.co/api/v1/plugs/" + DEVICE_ID + "/owners/" + OWNER_ID + "/consumption"

	csvFile, _ := os.Open(f)
	bread := bufio.NewReader(csvFile)
	reader := csv.NewReader(bread)

	var clist []Consumption

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(line)

		if len(line) != 10 {
			fmt.Println("Err: invalid data")
			break
		}

		//Convert data to slice
		P, _ := strconv.ParseFloat(line[9], 64)
		V := 0.0
		I := 0.0

		var T string

		dt := line[2]
		dt_strs := strings.Split(dt, " ")

		if len(dt_strs) == 2 {
			T = dt
		} else if len(dt_strs) == 1 {
			T = dt + " 00:00:00"
		} else {
			log.Fatal("Invalid datetime format: " + dt)
		}

		nc := Consumption{
			Timestamp:  T,
			Power:      P,
			Voltage:    V,
			Current:    I,
			Phaseshift: 0.0,
			Status:     1}

		clist = append(clist, nc)
	}

	for k, v := range clist {
		fmt.Println("line: ", k, v)
		if k != 0 && k >= l_num {
			//c := time.Tick(2 * time.Millisecond)
			//<-c
			for i := 0; i < RETRY_ON_FAIL; i++ {
				if sendToServer(&v) == true {
					break
				} else {
					fmt.Println("Retry send:", v)
				}
			}
		}
	}
}
