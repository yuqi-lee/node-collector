package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	mapName2ip   sync.Map
	mapName2veth sync.Map
)

const (
	csvPathIP   = "name_ip.csv"
	csvPathVeth = "name_veth.csv"
)

func podIPInfoInit() {
	err := genIPInfo()
	if err != nil {
		panic(err)
	}

	f, err := os.Open(csvPathIP)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	var record string
	for err != io.EOF {
		record, err = reader.Read()
		slice := strings.Split(record, " ")
		if len(slice) < 2 {
			log.Printf("csv file (name_ip.csv) format error.")
		} else {
			mapName2ip.Store(slice[0], slice[1])
		}
	}
}

func podVethInfoInit() {
	err := genVethInfo()
	if err != nil {
		panic(err)
	}

	f, err := os.Open(csvPathVeth)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	var record string
	for err != io.EOF {
		record, err = reader.Read()
		slice := strings.Split(record, " ")
		if len(slice) < 2 {
			log.Printf("csv file (name_veth.csv) format error.")
		} else {
			mapName2veth.Store(slice[0], slice[1])
		}
	}
}
func genIPInfo() error {
	cmd := exec.Command("kubectl", "get", "pods", "-n", "hotel")
	err := cmd.Run()

	return err
}

func genVethInfo() error {
	cmd := exec.Command("kubectl", "get", "pods", "-n", "hotel")
	err := cmd.Run()

	return err
}
