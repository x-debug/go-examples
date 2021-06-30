package main

import (
	"log"
	"time"

	"github.com/x-debug/timewheel/timex"
)

func main() {
	wheel := timex.NewTimeWheel(1, 60)
	wheel.SetTimer("test1", 2*time.Minute, func() {
		log.Println("hello, time wheel1")
	})
	wheel.SetTimer("test2", 1*time.Minute, func() {
		log.Println("hello, time wheel2")
	})
	wheel.SetTimer("test3", 39*time.Second, func() {
		log.Println("hello, time wheel3")
	})
	wheel.SetTimer("test4", 5*time.Minute, func() {
		log.Println("hello, time wheel4")
	})
	wheel.SetTimer("test5", 90*time.Second, func() {
		log.Println("hello, time wheel5")
	})
	wheel.SetTimer("test6", 3*time.Second, func() {
		log.Println("hello, time wheel6")
	})

	time.Sleep(30 * time.Minute)
	wheel.StopTimer()
}
