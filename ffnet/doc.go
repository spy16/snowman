// Package ffnet provides a generic Fully-Connected Feedforward neural network,
// and a Stochastic Gradient Descent based trainer.
//
// Following snippet shows how to create a simple 2-10-10 network with ReLU for
// hidden and Sigmoid for output layer activations and how to make a prediction.
//
//      net, _ := ffnet.New(2,
//          ffnet.Layer(10, ffnet.ReLU()),
//          ffnet.Layer(10, ffnet.Sigmoid()),
//      )
//      net.Predict(1, 1)
//
package ffnet
