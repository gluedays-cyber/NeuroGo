# NeuroGo (`.ngo`)

**NeuroGo** is an AI-augmented dialect and superset of the Go programming language designed for **Learned Control Flow** and **Intelligent Branching**.

Traditional programming languages rely strictly on deterministic branching (`if-else`, `switch-case`). NeuroGo extends Go syntax with native first-class primitives—`train`, `match`, and `score`—allowing developers to train neural network weights directly from datasets and route program execution using probabilistic confidence scores.

---

## Key Features

- **Native Intelligent Branching (`match ... score`)**: Replace hundreds of brittle heuristic rules with data-driven weight inference.
- **Offline Training Declaration (`train`)**: Declare dataset sources and training parameters directly inside your code, compiled into standalone `.gow` (Go Weight) archives.
- **Pure Go Inference Engine**: Zero external C/C++ dependencies (`CGO_ENABLED=0` friendly). Easily cross-compiles to a single static binary.
- **100% Go Interoperability**: NeuroGo transpiles directly into clean, idiomatic Go code. Seamlessly import and use any standard library or third-party Go package (`net/http`, `sync`, etc.).
- **Subword & Morphology Aware**: Built-in 2-gram / subword vectorization handles complex agglutinative languages (e.g., Korean, Japanese) as well as European languages out of the box.
- **Thread-Safe & Lock-Free**: In-memory immutable weight caches with zero-allocation buffers for massive Goroutine concurrency.

---

## Language Syntax Specification

### 1. `train` Block
Declares training parameters to construct a `.gow` model file prior to execution:

```go
train "intent_model.gow" {
    format: "csv",
    source: "dataset.csv",
    input:  "text",
    target: "label",
    epochs: 40,
}
```

### 2. `match` Statement
Evaluates an expression using the specified weight file and routes execution based on predicted class labels and confidence scores:

```go
match query using "intent_model.gow" {
case "Refund" score >= 0.70:
    routeToRefundAgent(query)
case "Delivery" score >= 0.70:
    trackShipment(query)
case "Inquiry" score >= 0.70:
    openProductFAQ(query)
default:
    fallbackSupport(query)
}
```

---

## Architecture Overview

```
[ .ngo Source Code ] ────▶ [ ngo Transpiler ] ────▶ [ Standard .go Code ]
        │                                                    │
   (train block)                                        (go run / build)
        ▼                                                    ▼
[ Training Pipeline ] ───▶ [ .gow Binary Archive ] ─▶ [ Pure Go Runtime ]
  (Vocab + SGD Trainer)      (Header + Weights)         (Forward Pass / Softmax)
```

### Generated Go Code Example
The transpiler cleanly rewrites `match` blocks into standard Go:

```go
{
    _ngoMatch := runtime.Match("intent_model.gow", query)
    switch {
    case _ngoMatch.Label == "Refund" && _ngoMatch.Score >= 0.70:
        routeToRefundAgent(query)
    case _ngoMatch.Label == "Delivery" && _ngoMatch.Score >= 0.70:
        trackShipment(query)
    default:
        fallbackSupport(query)
    }
}
```

---

## Project Structure

```text
neurogo/
├── cmd/
│   └── ngo/                   # 'ngo' CLI toolchain entry point
│       └── main.go
├── pkg/
│   ├── runtime/               # Model loading, tokenizer, and forward pass engine
│   │   └── runtime.go
│   ├── transpiler/            # Source-to-source code transformer (.ngo -> .go)
│   │   └── transpiler.go
│   └── trainer/               # CSV dataset parser, vocab builder, and SGD optimizer
│       └── trainer.go
├── examples/
│   └── intent/                # End-to-end customer intent routing demo
│       ├── dataset.csv        # Training samples
│       ├── main.ngo           # NeuroGo source file
│       ├── intent_model.gow   # Pre-trained weight artifact
│       └── main.go            # Transpiled Go file
├── neurogo_runner.py          # Standalone verification runner (Python 3 reference)
├── go.mod                     # Go module file
├── LICENSE                    # MIT License
└── README.md                  # Documentation and manual
```

---

## Getting Started

### Prerequisites
- **Go 1.20+** installed on your system.

### Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/gluedays-cyber/neurogo.git
   cd neurogo
   ```

2. Build the `ngo` CLI compiler:
   ```bash
   go build -o ngo ./cmd/ngo
   ```
   *(Optional)* Move `ngo` to your system PATH:
   ```bash
   sudo mv ngo /usr/local/bin/
   ```

---

## Usage Guide

### Step 1: Train the Model
Extract training declarations from your `.ngo` file and generate the `.gow` binary:
```bash
./ngo train examples/intent/main.ngo
```
Output:
```text
[NeuroGo Trainer] Starting training for 'intent_model.gow' from 'dataset.csv'...
[NeuroGo Trainer] Dataset parsed: 18 samples, 96 unique tokens, 3 classes: [Delivery Inquiry Refund]
   Epoch [ 10/ 40] - Loss: 0.0197
   Epoch [ 20/ 40] - Loss: 0.0100
   Epoch [ 30/ 40] - Loss: 0.0067
   Epoch [ 40/ 40] - Loss: 0.0051
[NeuroGo Trainer] Successfully saved model to 'intent_model.gow'!
```

### Step 2: Run Directly
Transpile and execute the code immediately:
```bash
./ngo run examples/intent/main.ngo
```

### Step 3: Inspect Transpiled Go Code
If you want to view the generated Go code:
```bash
./ngo transpile examples/intent/main.ngo -o examples/intent/main.go
```

### Step 4: Build a Native Binary
Compile into a standalone executable:
```bash
./ngo build examples/intent/main.ngo
go build -o myapp examples/intent/main.go
./myapp
```

---

## Verification & Standalone Simulation

If you do not have Go installed yet and wish to test the language pipeline immediately, run the Python 3 reference simulator:

```bash
python3 neurogo_runner.py
```

This runs the exact same parser, SGD trainer, and inference engine to verify transpilation, training, and branching logic.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Developed by [gluedays-cyber](https://github.com/gluedays-cyber).
