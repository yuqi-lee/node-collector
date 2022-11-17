package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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
)

func podIPInfoInit() {

	cmdStr := "kubectl --kubeconfig " + collectorConfig.KubeConfigPath + " get po -n " + collectorConfig.K8sNamespace + " -o wide |  awk '/" + collectorConfig.HostName + "/{print $1, $6}'"
	cmd := exec.Command("bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("kubectl get pods failed:%s\n", err.Error())
	}

	outStr /*errStr*/, _ := string(stdout.Bytes()), string(stderr.Bytes())
	// fmt.Printf("cmd: %s 's output is:\n %s\n", cmdStr, outStr)
	// fmt.Printf("cmd: %s 's error is:\n %s\n", cmdStr, errStr)
	outStr = strings.Replace(outStr, "\n", " ", -1)
	slice := strings.Split(outStr, " ")
	var i int
	for i = 0; i < len(slice)-1; i = i + 2 {
		mapName2ip.Store(slice[i], slice[i+1])
	}

	log.Printf("map (name -> ip) loads %d entries.", i/2)

	/*
		fmt.Printf("logs =================================\n")
		mapName2ip.Range(func(key, value interface{}) bool {
			name := key.(string)
			ip := value.(string)
			fmt.Printf("%s %s\n", name, ip)
			return true
		})
		fmt.Printf("logs =================================\n")
	*/
}

func genMap(outStr string) {
	outStr = strings.Replace(outStr, "\n", " ", -1)
	slice := strings.Split(outStr, " ")
	var i int
	for i = 0; i < len(slice)-1; i = i + 2 {
		mapName2ip.Store(slice[i], slice[i+1])
	}

	log.Printf("map (name -> ip) loads %d entries.", i/2)
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

func updateName2ip() {
	go func() {
		for {
			// 每隔 2 秒更新一次 map
			time.Sleep(2 * time.Second)

			//先清空所有的 Key
			mapName2ip.Range(func(k, v interface{}) bool {
				mapName2ip.Delete(k)
				return true
			})

			log.Printf("map (name -> ip) is empty now.")

			//重新 initialize
			podIPInfoInit()
		}
	}()
}
