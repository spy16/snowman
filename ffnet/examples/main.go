package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/spy16/snowman/ffnet"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// exampleXORNet()
	// exampleANDGate()

	net, err := ffnet.New(2, ffnet.Layer(10, ffnet.Sigmoid()))
	if err != nil {
		panic(err)
	}
	s := time.Now()
	fmt.Println(net.Predict(1, 0))
	log.Printf("predicted in %s", time.Since(s))
}
