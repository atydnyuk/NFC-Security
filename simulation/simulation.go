package main

import (
	"fmt"
	"time"
)

const (
	NUMBER_OF_DAYS = 1 //number of days to run simulation
	REQUESTS_PER_DAY = 100 //the popularity of the application
	MALICIOUS_USERS_FRACTION = .5 //fraction of the requests that are malicious

	ENABLE_DEDICATED_ATTACKERS = true //enables dedicated web attackers
	STRENGTH_OF_ATTACKERS = 5 //number of requests to bombard with when attack commences
	ATTACKS_PER_DAY = 5 //number of times a day to perform dedicated attacks

	//this specifies the granularity of the simulation. 
	//the higher the number, the more precise the results
	//and the longer the application takes to run
	SECONDS_PER_DAY = 24 
)
func main() {
	start := time.Now()
	fmt.Printf("Starting simulation at %d\n",start.Unix())
	//launch all of the routines
	go good_guys()
	go bad_guys()

	//wait until the correct time has elapsed
	seconds_needed := float64(NUMBER_OF_DAYS * SECONDS_PER_DAY)
	for (time.Since(start).Seconds() < seconds_needed) {
		time.Sleep(1)
	}
}

func good_guys() {
	for (true) {
		//fmt.Printf("This is a valid request\n")
		time.Sleep(1)
	}
}

func bad_guys() {
	for (true) {
		//fmt.Printf("This is a bad request\n")
		time.Sleep(1)
	}
}