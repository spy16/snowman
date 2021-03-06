package main

import (
	"context"
	"fmt"

	"github.com/spy16/snowman/pkg/ffnet"
)

func exampleANDGate() {
	n, err := ffnet.New(2,
		ffnet.Layer(5, ffnet.Sigmoid()), // ReLU hidden layer
		ffnet.Layer(1, ffnet.Sigmoid()), // sigmoid output layer
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(n)

	samples := []ffnet.Example{
		{[]float64{0, 0}, []float64{0}},
		{[]float64{0, 1}, []float64{0}},
		{[]float64{1, 0}, []float64{0}},
		{[]float64{1, 1}, []float64{1}},
	}

	fmt.Println("Before training:")
	for _, sample := range samples {
		yHat, _ := n.Predict(sample.Inputs...)
		fmt.Printf("for %v: %v\n", sample.Inputs, yHat)
	}

	trainer := ffnet.SGDTrainer{
		FFNet: n,
		Eta:   0.05,
		Loss:  ffnet.SquaredError(),
	}
	if err := trainer.Train(context.Background(), 5000, samples); err != nil {
		panic(err)
	}

	fmt.Println("After training:")
	for _, sample := range samples {
		yHat, _ := n.Predict(sample.Inputs...)
		fmt.Printf("for %v: %v\n", sample.Inputs, yHat)
	}
}
