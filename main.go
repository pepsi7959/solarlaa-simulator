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
	"time"
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

var URL string = "http://solaraa.advancedlogic.co/api/v1/plugs/112233445566/owners/2052/consumption"
var RETRY_ON_FAIL int = 5

func sendToServer(c *Consumption) bool {
	fmt.Println("send: ", c)
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	fmt.Println(c.Timestamp)
	jstr, err := json.Marshal(*c)

	if err != nil {
		log.Fatal(err)
		return false
	}

	resp, err := client.Post(URL, "application/json", bytes.NewBuffer(jstr))

	if err != nil {
		log.Fatal(err)
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

	if len(os.Args) == 1 {
		fmt.Println("err: Required csv filename")
	}

	f := os.Args[1]

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

		if len(line) != 5 {
			fmt.Println("Err: invalid data")
			break
		}

		//Convert data to slice
		P, _ := strconv.ParseFloat(line[2], 64)
		V, _ := strconv.ParseFloat(line[3], 64)
		I, _ := strconv.ParseFloat(line[4], 64)
		var T string
		if line[1] != "" && line[1] == "24.00" {
			d, _ := time.Parse("2006-01-02", line[0])
			d = d.AddDate(0, 0, 1)
			T = d.Format("2006-01-02") + " 00:00:00"
		} else {
			T = line[0] + " " + strings.Replace(line[1], ".", ":", 1) + ":00"
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

	for _, v := range clist {
		c := time.Tick(50 * time.Millisecond)
		<-c
		for i := 0; i < RETRY_ON_FAIL; i++ {
			if sendToServer(&v) == true {
				break
			} else {
				fmt.Println("Retry send:", v)
			}
		}
	}

}