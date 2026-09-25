package trainer

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"

	"neurogo/pkg/runtime"
	"neurogo/pkg/transpiler"
)

type Sample struct {
	Text   string
	Tokens []string
	Label  string
	Class  int
}

// TrainFromConfig trains a classification model based on a TrainConfig and outputs a .gow file.
func TrainFromConfig(cfg transpiler.TrainConfig) error {
	fmt.Printf("[NeuroGo Trainer] Starting training for '%s' from '%s'...\n", cfg.WeightPath, cfg.Source)

	f, err := os.Open(cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to open dataset source: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV dataset: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("dataset must contain header and at least one data row")
	}

	headerRow := records[0]
	inputColIdx := -1
	targetColIdx := -1

	for idx, col := range headerRow {
		colClean := strings.TrimSpace(col)
		if colClean == cfg.InputCol {
			inputColIdx = idx
		}
		if colClean == cfg.TargetCol {
			targetColIdx = idx
		}
	}

	if inputColIdx == -1 || targetColIdx == -1 {
		return fmt.Errorf("specified columns not found (input: %s, target: %s)", cfg.InputCol, cfg.TargetCol)
	}

	labelMap := make(map[string]int)
	var labels []string
	vocabMap := make(map[string]int)
	var samples []Sample

	// Build vocabulary and dataset samples
	for _, row := range records[1:] {
		if len(row) <= inputColIdx || len(row) <= targetColIdx {
			continue
		}
		text := row[inputColIdx]
		lbl := strings.TrimSpace(row[targetColIdx])

		classIdx, exists := labelMap[lbl]
		if !exists {
			classIdx = len(labels)
			labelMap[lbl] = classIdx
			labels = append(labels, lbl)
		}

		tokens := runtime.Tokenize(text)
		for _, token := range tokens {
			if _, ok := vocabMap[token]; !ok {
				vocabMap[token] = len(vocabMap)
			}
		}

		samples = append(samples, Sample{
			Text:   text,
			Tokens: tokens,
			Label:  lbl,
			Class:  classIdx,
		})
	}

	inputDim := len(vocabMap)
	numClasses := len(labels)

	if numClasses < 2 {
		return fmt.Errorf("need at least 2 distinct classes to train a classifier, found %d", numClasses)
	}

	fmt.Printf("[NeuroGo Trainer] Dataset parsed: %d samples, %d unique tokens, %d classes: %v\n",
		len(samples), inputDim, numClasses, labels)

	// Initialize weights and biases
	weights := make([]float32, inputDim*numClasses)
	biases := make([]float32, numClasses)
	for i := range weights {
		weights[i] = (rand.Float32() - 0.5) * 0.05
	}

	// SGD Training Loop
	learningRate := float32(0.1)
	epochs := cfg.Epochs
	if epochs <= 0 {
		epochs = 30
	}

	for epoch := 1; epoch <= epochs; epoch++ {
		totalLoss := 0.0

		// Shuffle samples each epoch
		rand.Shuffle(len(samples), func(i, j int) {
			samples[i], samples[j] = samples[j], samples[i]
		})

		for _, s := range samples {
			// Compute forward logits
			logits := make([]float64, numClasses)
			for c := 0; c < numClasses; c++ {
				sum := float64(biases[c])
				for _, tok := range s.Tokens {
					idx := vocabMap[tok]
					sum += float64(weights[idx*numClasses+c])
				}
				logits[c] = sum
			}

			// Softmax
			maxLogit := -math.MaxFloat64
			for _, l := range logits {
				if l > maxLogit {
					maxLogit = l
				}
			}
			var expSum float64
			probs := make([]float64, numClasses)
			for c := 0; c < numClasses; c++ {
				p := math.Exp(logits[c] - maxLogit)
				probs[c] = p
				expSum += p
			}
			for c := 0; c < numClasses; c++ {
				probs[c] /= expSum
			}

			// Loss = -log(probs[targetClass])
			targetProb := math.Max(probs[s.Class], 1e-12)
			totalLoss += -math.Log(targetProb)

			// Gradients and update
			// dL/d(logit_c) = probs[c] - y_c
			for c := 0; c < numClasses; c++ {
				y := 0.0
				if c == s.Class {
					y = 1.0
				}
				grad := float32(probs[c] - y)

				// Update bias
				biases[c] -= learningRate * grad

				// Update weights
				for _, tok := range s.Tokens {
					idx := vocabMap[tok]
					weights[idx*numClasses+c] -= learningRate * grad
				}
			}
		}

		if epoch%10 == 0 || epoch == epochs {
			avgLoss := totalLoss / float64(len(samples))
			fmt.Printf("   Epoch [%3d/%3d] - Loss: %.4f\n", epoch, epochs, avgLoss)
		}
	}

	header := runtime.Header{
		Magic:        runtime.MagicGOW,
		Version:      1,
		Architecture: "linear_softmax",
		InputDim:     inputDim,
		NumClasses:   numClasses,
		Labels:       labels,
		Vocab:        vocabMap,
	}

	if err := runtime.SaveGOW(cfg.WeightPath, header, weights, biases); err != nil {
		return fmt.Errorf("failed to save .gow file: %w", err)
	}

	fmt.Printf("[NeuroGo Trainer] Successfully saved model to '%s'!\n", cfg.WeightPath)
	return nil
}
