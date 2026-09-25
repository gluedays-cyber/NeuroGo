package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"neurogo/pkg/trainer"
	"neurogo/pkg/transpiler"
)

func printUsage() {
	fmt.Println("NeuroGo (ngo) - Intelligent Control Flow Language Toolchain")
	fmt.Println("\nUsage:")
	fmt.Println("  ngo train <file.ngo>              Train models defined in train blocks and generate .gow files")
	fmt.Println("  ngo transpile <file.ngo> [-o out] Transpile .ngo file to standard Go (.go)")
	fmt.Println("  ngo run <file.ngo>                Transpile and execute with 'go run'")
	fmt.Println("  ngo build <file.ngo> [-o binary]  Transpile and compile with 'go build'")
	fmt.Println("  ngo help                          Show this help message")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "help", "-h", "--help":
		printUsage()

	case "transpile":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to transpile")
			os.Exit(1)
		}
		ngoFile := os.Args[2]
		outFlag := flag.NewFlagSet("transpile", flag.ExitOnError)
		outFile := outFlag.String("o", "", "Output .go file path")
		outFlag.Parse(os.Args[3:])

		content, err := os.ReadFile(ngoFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		goCode, _, err := transpiler.Transpile(string(content))
		if err != nil {
			fmt.Printf("Transpilation failed: %v\n", err)
			os.Exit(1)
		}

		target := *outFile
		if target == "" {
			target = strings.TrimSuffix(ngoFile, filepath.Ext(ngoFile)) + ".go"
		}

		if err := os.WriteFile(target, []byte(goCode), 0644); err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[NeuroGo] Successfully transpiled '%s' -> '%s'\n", ngoFile, target)

	case "train":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to train")
			os.Exit(1)
		}
		ngoFile := os.Args[2]
		content, err := os.ReadFile(ngoFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		_, configs, err := transpiler.Transpile(string(content))
		if err != nil {
			fmt.Printf("Error parsing .ngo file: %v\n", err)
			os.Exit(1)
		}

		if len(configs) == 0 {
			fmt.Println("[NeuroGo] No 'train' blocks found in file.")
			return
		}

		for _, cfg := range configs {
			if err := trainer.TrainFromConfig(cfg); err != nil {
				fmt.Printf("Training error: %v\n", err)
				os.Exit(1)
			}
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to run")
			os.Exit(1)
		}
		ngoFile := os.Args[2]
		content, err := os.ReadFile(ngoFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		goCode, configs, err := transpiler.Transpile(string(content))
		if err != nil {
			fmt.Printf("Transpilation failed: %v\n", err)
			os.Exit(1)
		}

		// Auto train if weight file does not exist yet
		for _, cfg := range configs {
			if _, err := os.Stat(cfg.WeightPath); os.IsNotExist(err) {
				fmt.Printf("[NeuroGo] Weight file '%s' not found. Auto-training...\n", cfg.WeightPath)
				if err := trainer.TrainFromConfig(cfg); err != nil {
					fmt.Printf("Auto-training failed: %v\n", err)
					os.Exit(1)
				}
			}
		}

		targetGo := strings.TrimSuffix(ngoFile, filepath.Ext(ngoFile)) + ".go"
		if err := os.WriteFile(targetGo, []byte(goCode), 0644); err != nil {
			fmt.Printf("Error writing generated Go file: %v\n", err)
			os.Exit(1)
		}

		// Check if `go` is installed
		if _, err := exec.LookPath("go"); err != nil {
			fmt.Printf("[NeuroGo] Transpiled to '%s'. ('go' binary not found in current PATH to execute directly)\n", targetGo)
			return
		}

		cmd := exec.Command("go", "run", targetGo)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}

	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to build")
			os.Exit(1)
		}
		ngoFile := os.Args[2]
		content, err := os.ReadFile(ngoFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		goCode, _, err := transpiler.Transpile(string(content))
		if err != nil {
			fmt.Printf("Transpilation failed: %v\n", err)
			os.Exit(1)
		}

		targetGo := strings.TrimSuffix(ngoFile, filepath.Ext(ngoFile)) + ".go"
		if err := os.WriteFile(targetGo, []byte(goCode), 0644); err != nil {
			fmt.Printf("Error writing generated Go file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[NeuroGo] Transpiled to '%s'. Ready for 'go build'.\n", targetGo)

	default:
		fmt.Printf("Unknown command '%s'. Run 'ngo help' for usage.\n", command)
		os.Exit(1)
	}
}
