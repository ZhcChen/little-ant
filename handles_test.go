package main

import (
	"log"
	"testing"
	"time"
)

func TestGoFun(t *testing.T) {

	var a = 0

	go func() {
		_a := &a

		for {
			log.Println(" -------------------------- _a: ", *_a)
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)
		_a := &a
		*_a = 100
	}()

	time.Sleep(time.Second * 10)

}
