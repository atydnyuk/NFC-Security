package main

import (
	"fmt"
	"time"
	"math/rand"
)

const (
	NUMBER_OF_DAYS = 30 //number of days to run simulation
	REQUESTS_PER_DAY = 100 //the popularity of the application
	MALICIOUS_USERS_FRACTION = .1 //fraction of users that deface the tag
	NUMBER_OF_ATTACKERS = 10 //the malicious requests are split even
	
	//this specifies the granularity of the simulation. 
	//the higher the number, the more precise the results
	//and the longer the application takes to run
	SECONDS_PER_DAY = 24 
)

var rep_map map[int]int

func main() {
	rep_map = make(map[int]int)
		
	
	for i:= 0; i < NUMBER_OF_ATTACKERS; i++ {
		rep_map[i] = 0
	}

	start := time.Now()
	fmt.Printf("Starting simulation at %d\n",start.Unix())
	//launch all of the routines
	go good_guys(start)
	go bad_guys(start)

	//wait until the correct time has elapsed
	seconds_needed := float64(NUMBER_OF_DAYS * SECONDS_PER_DAY)
	for (time.Since(start).Seconds() < seconds_needed) {
		time.Sleep(1)
	}

	//then exit
}

/**
 *
 **/
func good_guys(t time.Time) {
	//this is to help the routnes start at around the same time
	time.Sleep(1*time.Nanosecond)
	good_per_day := REQUESTS_PER_DAY*(1-MALICIOUS_USERS_FRACTION)
	good_per_second := good_per_day/SECONDS_PER_DAY
	for (true) {
		//every tenth of a second
		time.Sleep(100 * time.Millisecond)
		odds := rand.Float64()
		if (odds < good_per_second/10) {
			fmt.Printf("Good request issued\n")
		}
		
	}
}

func bad_guys(t time.Time) {
	bad_per_day := REQUESTS_PER_DAY*(MALICIOUS_USERS_FRACTION)
	bad_per_second := bad_per_day/SECONDS_PER_DAY
	for (true) {
		//every tenth of a second
		time.Sleep(100 * time.Millisecond)
		odds := rand.Float64()
		if (odds < bad_per_second/10) {
			//we want to issue a malicious request from a bad guy
			badguy_id := rand.Intn(NUMBER_OF_ATTACKERS)
			rep_map[badguy_id] += 1
			if (rep_map[badguy_id] < 5) {
				fmt.Printf("Malicious request issued\n")
			}
		}
		
	}
}
