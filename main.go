package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var shasum = flag.String("shasumPath", "/usr/bin/shasum", "path to shasum binary")
var busRouteFile = flag.String("busRoutePath", "", "path to bus route file")
var routes = map[int32][]int32{}

func strToInt32(str string) (int32, error) {
	var err error
	var i64 int64
	if i64, err = strconv.ParseInt(str, 10, 32); err == nil {
		return int32(i64), err
	}
	return int32(0), err
}

func buildRouteMap(data []string) {
	busRouteId, _ := strToInt32(data[0])
	routes[busRouteId] = []int32{}
	for _, val := range data[1:] {
		stationId, _ := strToInt32(val)
		routes[busRouteId] = append(routes[busRouteId], stationId)
	}
}

func readFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Can't open file %s with error %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := strings.Split(scanner.Text(), " ")
		if len(str) > 1 {
			buildRouteMap(str)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func formatResp(depSid int32, arrSid int32, direct bool) string {
	resp := string(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"type": "object",
		"properties": {
			"dep_sid": {
				"type": "integer"
			},
			"arr_sid": {
				"type": "integer"
			},
			"direct_bus_route": {
				"type": "boolean"
			}
		},
		"required": [ "%v", "%v", "%v" ]
	}`)
	return fmt.Sprintf(resp, depSid, arrSid, direct)
}

func direct(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var depSid, arrSid int32
	var ok bool
	if _, ok = r.Form["dep_sid"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, formatResp(depSid, arrSid, false))
		return
	}
	if _, ok = r.Form["arr_sid"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, formatResp(depSid, arrSid, false))
		return
	}
	depSid, _ = strToInt32(strings.Join(r.Form["dep_sid"], ""))
	arrSid, _ = strToInt32(strings.Join(r.Form["arr_sid"], ""))

	dummySearch := func(data []int32, id int32, c chan int) {
		for key, val := range data {
			if val == id {
				c <- key
				return
			}
		}
		c <- -1
	}

	findRoute := func(data []int32, depSid int32, arrSid int32, done chan bool) {
		depChan := make(chan int)
		arrChan := make(chan int)
		go dummySearch(data, depSid, depChan)
		go dummySearch(data, arrSid, arrChan)
		depId := <-depChan
		arrId := <-arrChan
		if depId != -1 && arrId != -1 && depId < arrId {
			done <- true
			return
		}
		done <- false
		return
	}

	numBatch := 1000
	if len(routes) < 1000 {
		numBatch = len(routes) - 1
	}
	done := make(chan bool, numBatch)
	thisBatch := []int32{}
	count := 0
	for key := range routes {
		if count == numBatch {
			for _, val := range thisBatch {
				stations := routes[val]
				go findRoute(stations, depSid, arrSid, done)
			}

			for i := 0; i < numBatch; i++ {
				found := <-done
				if found {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, formatResp(depSid, arrSid, found))
					return
				}
			}
			count = 0
			thisBatch = []int32{}
		} else {
			thisBatch = append(thisBatch, key)
			count++
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, formatResp(depSid, arrSid, false))
}

func update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nUpdating  .....")
	readFile(*busRouteFile)
	fmt.Println("Ready .....")

	result, err := exec.Command(*shasum, *busRouteFile).Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Fail to update")
		return
	}
	output := strings.Split(string(result), " ")
	resp := string(`{"sha1sum": "%s"}`)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf(resp, output[0]))
}

func main() {

	flag.Parse()

	if *busRouteFile == "" {
		log.Fatal("Please provide route file")
	}
	fmt.Println("Initialization .....")
	readFile(*busRouteFile)
	fmt.Println("Ready .....")

	http.HandleFunc("/direct", direct)
	http.HandleFunc("/update", update)

	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
