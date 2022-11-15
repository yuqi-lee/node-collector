package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

var (
	mapName2ip   sync.Map
	mapName2veth sync.Map
)

const (
	csvPathIP   string = "name_ip.csv"
	csvPathVeth string = "name_veth.csv"

	dynamicConfig bool = true
)

func podIPInfoInit() {

	if dynamicConfig {
		err := genIPInfo()
		if err != nil {
			panic(err)
		}
	}

	f, err := os.Open(csvPathIP)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	var record []string
	for err != io.EOF {
		record, err = reader.Read()
		if len(record) < 2 {
			log.Printf("csv file (name_ip.csv) format error with %d colum.", len(record))
		} else {
			mapName2ip.Store(record[0], record[1])
		}
	}
}

func podVethInfoInit() {

	/*
		if dynamicConfig {
			err := genVethInfo()
			if err != nil {
				panic(err)
			}
		}*/

	f, err := os.Open(csvPathVeth)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	var record []string
	for err != io.EOF {
		record, err = reader.Read()
		if len(record) < 2 {
			log.Printf("csv file (name_ip.csv) format error.")
		} else {
			mapName2ip.Store(record[0], record[1])
		}
	}
}
func genIPInfo() error {
	cmd := exec.Command("bash", "-c", "kubectl get po -n hotel-reservation -o wide |  awk '/skv-node4/{print $1, $6}' > name_ip.csv")
	err := cmd.Run()
	log.Println(err)
	return err
}

func genVethInfo() error {

	return nil
}

func updateName2ip() {
	go func() {
		for {
			// 每隔一秒更新一次map
			time.Sleep(time.Second)

			//先清空所有的 Key
			mapName2ip.Range(func(k, v interface{}) bool {
				mapName2ip.Delete(k)
				return true
			})

			//重新 initialize
			podIPInfoInit()
		}
	}()
}
