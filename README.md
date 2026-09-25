# NeuroGo (`.ngo`)

**Beyond `if`/`switch`: An AI-augmented dialect of Go powered by `match` and `score`.**

Traditional programming languages rely strictly on deterministic branching (`if-else`, `switch-case`)[cite: 1]. **NeuroGo** extends Go syntax with native intelligent branching—`match` and `score`—allowing developers to route program execution using probabilistic confidence scores driven by offline pre-trained neural network weights[cite: 1].

---

## Motivation: The Missing Middle in Control Flow

Modern software development is polarized between two extreme approaches to decision-making:

1. **Deterministic Branching (`if`/`switch`)**: Extremely fast, predictable, and lightweight, but fundamentally brittle. Handling natural language or unstructured input requires thousands of heuristic rules, regex patterns, or string checks that fail on minor typos, morphological variations, or synonyms.
2. **Heavyweight LLMs (Large Language Models)**: Highly flexible and semantically rich, but heavily over-engineered for simple control flow. LLMs introduce massive memory footprints (gigabytes of VRAM), network latency (hundreds of milliseconds), monetary API costs, and non-deterministic behavior.

```
Rigid Heuristics ──────────────▶ [ NeuroGo: The Sweet Spot ] ──────────────▶ Heavyweight LLMs
(Fast, Zero Tolerance)            (Sub-millisecond, Flexible, Pure Go)        (Slow, Expensive, Overkill)
```

### Bridging the Gap

NeuroGo introduces a third scope: **lightweight, data-driven statistical branching**.

- **Flexibility without Bloat**: Maps inputs into a continuous semantic latent space rather than relying on discrete string matches, absorbing natural language variances without external dependencies.
- **Microsecond In-Memory Execution**: Operates locally in pure Go (`CGO_ENABLED=0`) with kilobyte-scale weights, achieving sub-millisecond routing speeds on standard CPU cores.
- **Threshold-Driven Rejection**: Incorporates confidence scores (`score >= 0.70`) directly into language primitives, offering a formal probabilistic fallback mechanism (`default`) when input certainty is low.

By injecting bounded statistical flexibility into Go's strict control flow, NeuroGo delivers resilient semantic routing without the architectural baggage of modern deep learning stacks.

---

## Architectural Principle: Separation of Train and Match

In modern software engineering and MLOps, **model training (data science/optimization)** and **runtime serving (application control flow)** have fundamentally different lifecycles:

```
[ Model Pipeline (Offline / MLOps) ]
  dataset.csv ────▶ [ ngotrain CLI ] ────▶ model.gow (Immutable Weights)

[ Application Pipeline (Runtime / Serving) ]
  main.ngo    ────▶ [ ngo CLI ]      ────▶ main.go ────▶ Binary Execution
                         ▲                   │
                         └── (reads model.gow)
```

1. **`ngotrain` (Model Training Pipeline)**:
   - Dedicated offline CLI tool for dataset parsing, vocabulary extraction, and SGD optimization.
   - Independent of the application compiler; retrain and update `.gow` weight files without touching or rebuilding the application code.
2. **`ngo` (Application Toolchain & Intelligent Branching)**:
   - Pure transpiler and build orchestrator.
   - Converts `match ... score` syntax into idiomatic Go code linked against the pure Go inference runtime.
   - Eliminates compilation overhead by completely removing training logic and heavy dataset scans from the application build step.

---

## Key Features

- **Native Intelligent Branching (`match ... score`)**: Replace hundreds of brittle heuristic rules with data-driven weight inference[cite: 1].
- **Decoupled Architecture**: Clear boundary between offline training (`ngotrain`) and application runtime (`ngo`).
- **Pure Go Inference Engine**: Zero external C/C++ dependencies (`CGO_ENABLED=0` friendly)[cite: 1]. Easily cross-compiles to a single static binary[cite: 1].
- **100% Go Interoperability**: NeuroGo transpiles directly into clean, idiomatic Go code[cite: 1]. Seamlessly import and use any standard library or third-party Go package (`net/http`, `sync`, etc.)[cite: 1].
- **Subword & Morphology Aware**: Built-in 2-gram / subword vectorization handles typos and linguistic variations out of the box[cite: 1].
- **Thread-Safe & Lock-Free**: In-memory immutable weight caches with zero-allocation buffers for massive Goroutine concurrency[cite: 1].

---

## Language Syntax Specification

### `match` Statement
Evaluates an expression using the specified weight file and routes execution based on predicted class labels and confidence scores[cite: 1]:

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

## Project Structure

> **Working Directory Rule**: Always run terminal commands from the **root directory** of this repository (`NeuroGo/`)[cite: 1].

```text
NeuroGo/                       <-- Run all commands from this root folder
├── cmd/
│   ├── ngo/                   # 'ngo' Application CLI (transpile, run, build)
│   │   └── main.go
│   └── ngotrain/              # 'ngotrain' Offline Training CLI
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
│       ├── dataset.csv        # Training samples
│       ├── main.ngo           # NeuroGo source file
│       ├── intent_model.gow   # Pre-trained weight artifact
│       └── main.go            # Generated Go file
├── neurogo_runner.py          # Standalone verification runner (Python 3 reference)
├── go.mod                     # Go module file
├── LICENSE                    # MIT License
└── README.md                  # Documentation and manual
```

---

## Step-by-Step Quickstart Guide

### Prerequisites
- **Git** installed on your system (`git --version`)[cite: 1].
- **Go 1.20+** installed on your system (`go version`)[cite: 1].

---

### Step 0: Clone the Repository and Navigate to the Directory

```bash
git clone [https://github.com/gluedays-cyber/NeuroGo.git](https://github.com/gluedays-cyber/NeuroGo.git)
cd NeuroGo
```

---

### Step 1: Install Toolchains

Install both the compiler toolchain (`ngo`) and the training tool (`ngotrain`)[cite: 1]:

```bash
go install ./cmd/ngo
go install ./cmd/ngotrain
```

---

### Step 2: Train the Model Offline (`ngotrain`)

Train a classification model from your dataset to produce the binary weight file (`.gow`):

```bash
ngotrain -data examples/intent/dataset.csv -out examples/intent/intent_model.gow -epochs 40
```

**What happens?**
The `ngotrain` utility parses `dataset.csv`, extracts the vocabulary, optimizes weights via SGD cross-entropy, and outputs the standalone binary weight archive to `intent_model.gow`.

**Expected Output:**
```text
[ngotrain] Starting training on 'examples/intent/dataset.csv' -> 'examples/intent/intent_model.gow' (epochs: 40)...
[NeuroGo Trainer] Starting training for 'examples/intent/intent_model.gow' from 'examples/intent/dataset.csv'...
[NeuroGo Trainer] Dataset parsed: 18 samples, 195 unique tokens, 3 classes: [Refund Delivery Inquiry]
   Epoch [ 10/ 40] - Loss: 0.0191
   Epoch [ 20/ 40] - Loss: 0.0098
   Epoch [ 30/ 40] - Loss: 0.0066
   Epoch [ 40/ 40] - Loss: 0.0050
[NeuroGo Trainer] Successfully saved model to 'examples/intent/intent_model.gow'!
[ngotrain] Successfully generated model weights: 'examples/intent/intent_model.gow'
```

---

### Step 3: Run Intelligent Branching (`ngo run`)

Execute the program with real-time AI-based routing[cite: 1]:

```bash
ngo run examples/intent/main.ngo
```

**What happens?**
1. Transpiles `main.ngo` into valid standard Go code (`main.go`)[cite: 1].
2. Evaluates the test queries against the pre-trained weights using pure Go tensor operations[cite: 1].
3. Dynamically branches to the matching `case` based on the predicted class and confidence score[cite: 1].

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

---

### Step 4: Inspect Generated Standard Go Code

To inspect the generated code without immediate execution[cite: 1]:

```bash
ngo transpile examples/intent/main.ngo -o examples/intent/main.go
```

Examine how NeuroGo maps `match` statements to standard Go control flow in `examples/intent/main.go`[cite: 1]:

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

To verify the tokenizer, SGD loop, and branching logic end-to-end without compiling[cite: 1]:

```bash
# Windows
python neurogo_runner.py

# macOS / Linux
python3 neurogo_runner.py
```

---

## Frequently Asked Questions (FAQ)

### 1. `ngo: command not found` or `'ngo' is not recognized`
Verify your Go binary directory is included in your system `PATH`[cite: 1]:
- **Windows**: Add `%USERPROFILE%\go\bin` to `PATH`[cite: 1].
- **macOS / Linux**: Add `export PATH=$PATH:$(go env GOPATH)/bin` to your `~/.bashrc` or `~/.zshrc`[cite: 1].

### 2. Can I update models without recompiling the application?
Yes. Because `ngotrain` is strictly separated from `ngo`, you can retrain on updated datasets and overwrite `.gow` weight files at any time. The application runtime automatically picks up the updated weights without source modification.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details[cite: 1].

Developed by [gluedays-cyber](https://github.com/gluedays-cyber)[cite: 1].
