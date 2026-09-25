package main

import (
	"flag"
	"fmt"
	"os"

	"neurogo/pkg/trainer"
	"neurogo/pkg/transpiler"
)

func main() {
	sourceFlag := flag.String("data", "", "Path to dataset CSV file (required)")
	outputFlag := flag.String("out", "model.gow", "Output weight file path (.gow)")
	inputCol := flag.String("input", "text", "Input column name in CSV")
	targetCol := flag.String("target", "label", "Target column name in CSV")
	epochs := flag.Int("epochs", 40, "Number of training epochs")
	flag.Parse()

	if *sourceFlag == "" {
		fmt.Println("Error: -data flag is required.")
		fmt.Println("Usage: ngotrain -data dataset.csv -out intent_model.gow [-epochs 40]")
		os.Exit(1)
	}

	cfg := transpiler.TrainConfig{
		Format:     "csv",
		Source:     *sourceFlag,
		InputCol:   *inputCol,
		TargetCol:  *targetCol,
		WeightPath: *outputFlag,
		Epochs:     *epochs,
	}

	fmt.Printf("[ngotrain] Starting training on '%s' -> '%s' (epochs: %d)...\n", cfg.Source, cfg.WeightPath, cfg.Epochs)
	if err := trainer.TrainFromConfig(cfg); err != nil {
		fmt.Printf("[ngotrain] Training failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[ngotrain] Successfully generated model weights: '%s'\n", cfg.WeightPath)
}
