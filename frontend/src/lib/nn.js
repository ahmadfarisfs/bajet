// Minimal feedforward neural network — sigmoid activation, MSE loss, SGD backprop.
// Drop-in replacement for brain.js NeuralNetwork with the same object-key API.
// Zero native dependencies; runs identically in Node and browsers.

function sigmoid(x)      { return 1 / (1 + Math.exp(-x)) }
function sigmoidDeriv(x) { const s = sigmoid(x); return s * (1 - s) }
function rand()          { return (Math.random() * 2 - 1) * 0.5 }

export class NeuralNetwork {
  constructor({ hiddenLayers = [8, 4] } = {}) {
    this._hidden = hiddenLayers
    this._layers = null
    this._inKeys = null
    this._outKeys = null
  }

  _build(inSize, outSize) {
    const sizes = [inSize, ...this._hidden, outSize]
    this._layers = []
    for (let i = 1; i < sizes.length; i++) {
      const ni = sizes[i - 1], no = sizes[i]
      this._layers.push({
        w: Array.from({ length: no }, () => Array.from({ length: ni }, rand)),
        b: Array.from({ length: no }, rand),
        ni, no,
      })
    }
  }

  _forward(inp) {
    const acts = [inp]
    const zs   = []
    for (const L of this._layers) {
      const z = L.w.map((row, j) =>
        row.reduce((s, w, k) => s + w * acts[acts.length - 1][k], L.b[j])
      )
      zs.push(z)
      acts.push(z.map(sigmoid))
    }
    return { acts, zs }
  }

  train(data, { iterations = 2000, errorThresh = 0.005, learningRate = 0.3 } = {}) {
    this._inKeys  = Object.keys(data[0].input)
    this._outKeys = Object.keys(data[0].output)
    this._build(this._inKeys.length, this._outKeys.length)

    for (let iter = 0; iter < iterations; iter++) {
      let totalErr = 0
      for (const sample of data) {
        const inp    = this._inKeys.map(k => sample.input[k] ?? 0)
        const target = this._outKeys.map(k => sample.output[k] ?? 0)

        const { acts, zs } = this._forward(inp)
        const out = acts[acts.length - 1]
        totalErr += target.reduce((s, t, i) => s + (t - out[i]) ** 2, 0) / target.length

        // Output layer delta
        let delta = out.map((o, i) => (o - target[i]) * sigmoidDeriv(zs[zs.length - 1][i]))

        // Backprop through layers (reverse)
        for (let l = this._layers.length - 1; l >= 0; l--) {
          const L    = this._layers[l]
          const prev = acts[l]

          for (let j = 0; j < L.no; j++) {
            L.b[j] -= learningRate * delta[j]
            for (let k = 0; k < L.ni; k++) L.w[j][k] -= learningRate * delta[j] * prev[k]
          }

          if (l > 0) {
            delta = prev.map((_, k) => {
              const sum = L.w.reduce((s, row, j) => s + row[k] * delta[j], 0)
              return sum * sigmoidDeriv(zs[l - 1][k])
            })
          }
        }
      }
      if (totalErr / data.length < errorThresh) break
    }
  }

  run(input) {
    const inp  = this._inKeys.map(k => input[k] ?? 0)
    const { acts } = this._forward(inp)
    const out  = acts[acts.length - 1]
    return Object.fromEntries(this._outKeys.map((k, i) => [k, out[i]]))
  }
}
