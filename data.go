package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

const numRoutes = 100000
const numStations = 1000000
const stationPerRoute = 1000

func main() {
	var file *os.File
	var err error
	if file, err = os.Create("data.txt"); err != nil {
		panic(fmt.Sprintf("fail to create data file data.txt", "data.txt"))
		return
	}
	defer file.Close()
	for id := 0; id < numRoutes; id++ {
		stations := rand.Perm(numStations)[id : id+stationPerRoute]

		routes := strconv.Itoa(id) + " "
		for key, val := range stations {
			routes += strconv.Itoa(val)
			if key == len(stations)-1 {
				routes += "\n"
			} else {
				routes += " "
			}
		}

		if _, err := file.WriteString(routes); err != nil {
			panic(fmt.Sprintf("fail to write data to file %s", "data.txt"))
			return
		}
	}
}
