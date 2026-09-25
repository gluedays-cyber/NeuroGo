package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"neurogo/pkg/transpiler"
)

func printUsage() {
	fmt.Println("NeuroGo (ngo) - Intelligent Control Flow Language Toolchain")
	fmt.Println("\nUsage:")
	fmt.Println("  ngo transpile <file.ngo> [-o out]  Transpile .ngo file to standard Go (.go)")
	fmt.Println("  ngo run <file.ngo>                 Transpile and execute with 'go run'")
	fmt.Println("  ngo build <file.ngo> [-o binary]   Transpile and compile with 'go build'")
	fmt.Println("  ngo help                           Show this help message")
}

func loadAndTranspile(ngoPath string) (string, error) {
	absNgoPath, err := filepath.Abs(ngoPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for '%s': %w", ngoPath, err)
	}

	content, err := os.ReadFile(absNgoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", absNgoPath, err)
	}

	goCode, _, err := transpiler.Transpile(string(content))
	if err != nil {
		return "", fmt.Errorf("transpilation failed: %w", err)
	}

	return goCode, nil
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

		goCode, err := loadAndTranspile(ngoFile)
		if err != nil {
			fmt.Println(err)
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

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to run")
			os.Exit(1)
		}
		ngoFile := os.Args[2]

		goCode, err := loadAndTranspile(ngoFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		targetGo := strings.TrimSuffix(ngoFile, filepath.Ext(ngoFile)) + ".go"
		if err := os.WriteFile(targetGo, []byte(goCode), 0644); err != nil {
			fmt.Printf("Error writing generated Go file: %v\n", err)
			os.Exit(1)
		}

		if _, err := exec.LookPath("go"); err != nil {
			fmt.Printf("[NeuroGo] Transpiled to '%s'. ('go' binary not found in current PATH)\n", targetGo)
			return
		}

		absTargetGo, err := filepath.Abs(targetGo)
		if err != nil {
			fmt.Printf("Error resolving target path: %v\n", err)
			os.Exit(1)
		}

		cmd := exec.Command("go", "run", filepath.Base(absTargetGo))
		cmd.Dir = filepath.Dir(absTargetGo)
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
		buildFlag := flag.NewFlagSet("build", flag.ExitOnError)
		outFile := buildFlag.String("o", "", "Output binary file path")
		buildFlag.Parse(os.Args[3:])

		goCode, err := loadAndTranspile(ngoFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		targetGo := strings.TrimSuffix(ngoFile, filepath.Ext(ngoFile)) + ".go"
		if err := os.WriteFile(targetGo, []byte(goCode), 0644); err != nil {
			fmt.Printf("Error writing generated Go file: %v\n", err)
			os.Exit(1)
		}

		if _, err := exec.LookPath("go"); err != nil {
			fmt.Printf("[NeuroGo] Transpiled to '%s'. ('go' binary not found in current PATH)\n", targetGo)
			return
		}

		absTargetGo, err := filepath.Abs(targetGo)
		if err != nil {
			fmt.Printf("Error resolving target path: %v\n", err)
			os.Exit(1)
		}

		var args []string
		args = append(args, "build")
		if *outFile != "" {
			absOut, err := filepath.Abs(*outFile)
			if err == nil {
				args = append(args, "-o", absOut)
			} else {
				args = append(args, "-o", *outFile)
			}
		}
		args = append(args, filepath.Base(absTargetGo))

		cmd := exec.Command("go", args...)
		cmd.Dir = filepath.Dir(absTargetGo)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}

		fmt.Printf("[NeuroGo] Successfully built binary from '%s'\n", targetGo)

	default:
		fmt.Printf("Unknown command '%s'. Run 'ngo help' for usage.\n", command)
		os.Exit(1)
	}
}
