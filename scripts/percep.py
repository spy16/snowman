import numpy as np


def sigmoid(x):
    return 1.0 / (1.0 + np.exp(-x))


def sigmoid_prime(y_hat):
    return y_hat * (1 - y_hat)


class Perceptron:
    """Perceptron implements a simple perceptron cell."""

    def __init__(self, inputs):
        self.weights = np.random.randn(1, inputs)
        self.bias = 0

    def predict(self, inputs):
        z = np.dot(self.weights, inputs) + self.bias
        return sigmoid(z)

    def train(self, x, y, epochs=100, eta=1):
        for i in range(epochs):
            # 1. make a prediction for given training input
            y_hat = self.predict(x)
            print("loss={}".format((1. / 2.) * np.power(y - y_hat, 2)))

            # 2. estimate error (delta)
            delta = (y_hat - y) * sigmoid_prime(y_hat)

            # 3. calculate adjustments for weights and bias
            dW = np.dot(delta, x.T)
            dB = delta

            # 4. update weights and bias
            self.weights = self.weights - eta * dW
            self.bias = self.bias - eta * dB


if __name__ == "__main__":
    p = Perceptron(2)

    # OR gate input combinations and outputs
    x = np.array([[0, 0],
                  [0, 1],
                  [1, 0],
                  [1, 1]]).T
    y = np.array([[0, 1, 1, 1]])

    print("Before Training:\n{} -> {}".format(x.T, p.predict(x)))
    p.train(x, y)
    print("After Training:\n{} -> {}".format(x.T, p.predict(x)))
