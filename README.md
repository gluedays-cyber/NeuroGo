# NeuroGo (`.ngo`)

**Beyond `if`/`switch`: An AI-augmented dialect of Go powered by `train`, `match`, and `score`.**

Traditional programming languages rely strictly on deterministic branching (`if-else`, `switch-case`). **NeuroGo** extends Go syntax with native first-class primitives—`train`, `match`, and `score`—allowing developers to train neural network weights directly from datasets and route program execution using probabilistic confidence scores.

---

## Key Features

- **Native Intelligent Branching (`match ... score`)**: Replace hundreds of brittle heuristic rules with data-driven weight inference.
- **Offline Training Declaration (`train`)**: Declare dataset sources and training parameters directly inside your code, compiled into standalone `.gow` (Go Weight) archives.
- **Pure Go Inference Engine**: Zero external C/C++ dependencies (`CGO_ENABLED=0` friendly). Easily cross-compiles to a single static binary.
- **100% Go Interoperability**: NeuroGo transpiles directly into clean, idiomatic Go code. Seamlessly import and use any standard library or third-party Go package (`net/http`, `sync`, etc.).
- **Subword & Morphology Aware**: Built-in 2-gram / subword vectorization handles complex morphology and typos out of the box.
- **Thread-Safe & Lock-Free**: In-memory immutable weight caches with zero-allocation buffers for massive Goroutine concurrency.

---

## Language Syntax Specification

### 1. `train` Block
Declares dataset sources and training parameters to construct a `.gow` model file:

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

---

## Project Structure

> **Working Directory Rule**: Always run terminal commands from the **root directory** of this repository (`NeuroGo/`).

```text
NeuroGo/                       <-- Run all commands from this root folder
├── cmd/
│   └── ngo/                   # 'ngo' CLI compiler source code
│       └── main.go
├── pkg/
│   ├── runtime/               # Model loading, tokenizer, and forward pass engine
│   │   └── runtime.go
│   ├── transpiler/            # Source-to-source code transformer (.ngo -> .go)
│   │   └── transpiler.go
│   └── trainer/               # CSV dataset parser, vocab builder, and SGD optimizer
│       └── trainer.go
├── examples/
│   └── intent/                # Customer intent routing example
│       ├── dataset.csv        # English training samples
│       ├── main.ngo           # NeuroGo source file
│       ├── intent_model.gow   # Pre-trained weight artifact
│       └── main.go            # Transpiled Go file
├── neurogo_runner.py          # Standalone verification runner (Python 3 reference)
├── go.mod                     # Go module file
├── LICENSE                    # MIT License
└── README.md                  # Documentation and manual
```

---

## Step-by-Step Quickstart Guide

### Prerequisites
- **Go 1.20+** installed on your system. Verify by running `go version` in your terminal.

---

### Step 1: Install the `ngo` Compiler Tool (First Time Only)

From the root directory of the repository, install the `ngo` command:

```bash
go install ./cmd/ngo
```

**Why `go install`?**
- It automatically handles platform executable formats (`ngo.exe` on Windows, `ngo` on macOS/Linux).
- It installs the binary to your Go bin directory (`~/go/bin` or `%GOPATH%\bin`), allowing you to run `ngo` from **any directory** without worrying about `.\`, `./`, or `.exe` extension issues.

*(Alternative: If you prefer building locally in the current folder, run `go build ./cmd/ngo` without `-o`. On Windows this creates `ngo.exe`; on macOS/Linux it creates `ngo`.)*

---

### Step 2: Train the AI Model (`train`)

Run the trainer on the example NeuroGo program:

```bash
# Using installed tool (Recommended)
ngo train examples/intent/main.ngo

# Or using local binary:
# On Windows:      .\ngo.exe train examples/intent/main.ngo
# On macOS/Linux:  ./ngo train examples/intent/main.ngo
```

**What happens?**
The tool parses the `train` block in `main.ngo`, reads `dataset.csv`, trains a classification model using SGD cross-entropy optimization, and saves the binary weights to `examples/intent/intent_model.gow`.

**Expected Output:**
```text
[NeuroGo Trainer] Starting training for 'examples/intent/intent_model.gow' from 'examples/intent/dataset.csv'...
[NeuroGo Trainer] Dataset parsed: 18 samples, 195 unique tokens, 3 classes: [Delivery Inquiry Refund]
   Epoch [ 10/ 40] - Loss: 0.0050
   Epoch [ 20/ 40] - Loss: 0.0032
   Epoch [ 30/ 40] - Loss: 0.0024
   Epoch [ 40/ 40] - Loss: 0.0019
[NeuroGo Trainer] Successfully saved model to 'examples/intent/intent_model.gow'!
```

---

### Step 3: Run Intelligent Branching (`run`)

Execute the program with real-time AI-based routing:

```bash
# Using installed tool (Recommended)
ngo run examples/intent/main.ngo

# Or using local binary:
# On Windows:      .\ngo.exe run examples/intent/main.ngo
# On macOS/Linux:  ./ngo run examples/intent/main.ngo
```

**What happens?**
1. Transpiles `main.ngo` into valid standard Go code (`main.go`).
2. Evaluates the test queries against the trained weights using pure Go tensor operations.
3. Dynamically branches to the matching `case` based on the predicted class and confidence score.

**Expected Output:**
```text
=== NeuroGo Intelligent Branching Demo ===

[User Input] Please cancel my order and issue a full refund
>> [Routing] Connecting to Refund & Billing Specialist...

[User Input] Where is my package? Track shipment please
>> [Routing] Launching Real-time Shipment Tracking...

[User Input] Do you have this jacket in size medium?
>> [Routing] Directing to Product FAQ & Support Bot...

[User Input] What should I have for lunch today?
>> [Routing] Low confidence query. Routing to General Helpdesk.
```

Notice that the out-of-domain query (*"What should I have for lunch today?"*) fails to meet the 70% confidence threshold (`score >= 0.70`) and is safely routed to the `default` fallback branch.

---

### Step 4: Inspect Generated Standard Go Code

To see how NeuroGo transforms `match` statements into idiomatic Go code without executing it:

```bash
ngo transpile examples/intent/main.ngo -o examples/intent/main.go
```

Open `examples/intent/main.go` to examine the transpiled `switch` block:

```go
{
    _ngoMatch := runtime.Match("intent_model.gow", query)
    switch {
    case _ngoMatch.Label == "Refund" && _ngoMatch.Score >= 0.70:
        fmt.Println(">> [Routing] Connecting to Refund & Billing Specialist...")
    case _ngoMatch.Label == "Delivery" && _ngoMatch.Score >= 0.70:
        fmt.Println(">> [Routing] Launching Real-time Shipment Tracking...")
    case _ngoMatch.Label == "Inquiry" && _ngoMatch.Score >= 0.70:
        fmt.Println(">> [Routing] Directing to Product FAQ & Support Bot...")
    default:
        fmt.Println(">> [Routing] Low confidence query. Routing to General Helpdesk.")
    }
}
```

---

## Standalone Python Simulation (No Go Compiler Required)

If you wish to test the entire pipeline without configuring a Go development environment, run the bundled reference simulator from the root directory:

```bash
# Windows
python neurogo_runner.py

# macOS / Linux
python3 neurogo_runner.py
```

This simulates the exact tokenizer, SGD training loop, and inference branching logic end-to-end.

---

## Frequently Asked Questions (FAQ)

### 1. `ngo: command not found` or `'ngo' is not recognized`
If you ran `go install ./cmd/ngo` but your terminal cannot find `ngo`, make sure your Go bin path is in your system's `PATH` environment variable:
- **Windows**: Add `%USERPROFILE%\go\bin` to your `PATH`.
- **macOS / Linux**: Add `export PATH=$PATH:$(go env GOPATH)/bin` to your `~/.bashrc` or `~/.zshrc`.
- Alternatively, build locally in your project folder with `go build ./cmd/ngo` and run `.\ngo.exe` (Windows) or `./ngo` (macOS/Linux).

### 2. Why avoid `go build -o ngo ./cmd/ngo` on Windows?
On Windows, passing `-o ngo` forces the compiler to create an extensionless file named `ngo` instead of `ngo.exe`. Windows cannot execute binary files without the `.exe` extension. Running `go install ./cmd/ngo` or `go build ./cmd/ngo` (without `-o`) automatically generates the proper `.exe` extension on Windows.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Developed by [gluedays-cyber](https://github.com/gluedays-cyber).
