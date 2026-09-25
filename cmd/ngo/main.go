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
	fmt.Println("  ngo train <file.ngo>               Train models defined in train blocks and generate .gow files")
	fmt.Println("  ngo transpile <file.ngo> [-o out]  Transpile .ngo file to standard Go (.go)")
	fmt.Println("  ngo run <file.ngo>                 Transpile and execute with 'go run'")
	fmt.Println("  ngo build <file.ngo> [-o binary]   Transpile and compile with 'go build'")
	fmt.Println("  ngo help                           Show this help message")
}

// loadAndTranspile: 공통 파일 로딩, 트랜스파일, 경로 보정 수행
func loadAndTranspile(ngoPath string) (string, []transpiler.TrainConfig, error) {
	absNgoPath, err := filepath.Abs(ngoPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve absolute path for '%s': %w", ngoPath, err)
	}

	content, err := os.ReadFile(absNgoPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file '%s': %w", absNgoPath, err)
	}

	goCode, configs, err := transpiler.Transpile(string(content))
	if err != nil {
		return "", nil, fmt.Errorf("transpilation failed: %w", err)
	}

	// .ngo 파일이 위치한 디렉터리를 기준으로 상대 경로 보정
	baseDir := filepath.Dir(absNgoPath)
	for i := range configs {
		if !filepath.IsAbs(configs[i].Source) {
			configs[i].Source = filepath.Join(baseDir, configs[i].Source)
		}
		if !filepath.IsAbs(configs[i].WeightPath) {
			configs[i].WeightPath = filepath.Join(baseDir, configs[i].WeightPath)
		}
	}

	return goCode, configs, nil
}

// executeTraining: 보정된 설정값 기반으로 모델 학습 실행
func executeTraining(configs []transpiler.TrainConfig) error {
	for _, cfg := range configs {
		if err := trainer.TrainFromConfig(cfg); err != nil {
			return fmt.Errorf("training error for '%s': %w", cfg.WeightPath, err)
		}
	}
	return nil
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

		goCode, _, err := loadAndTranspile(ngoFile)
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

	case "train":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to train")
			os.Exit(1)
		}
		ngoFile := os.Args[2]

		_, configs, err := loadAndTranspile(ngoFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(configs) == 0 {
			fmt.Println("[NeuroGo] No 'train' blocks found in file.")
			return
		}

		if err := executeTraining(configs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a .ngo file to run")
			os.Exit(1)
		}
		ngoFile := os.Args[2]

		goCode, configs, err := loadAndTranspile(ngoFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 가중치 파일 미존재 시 자동 학습 수행
		for _, cfg := range configs {
			if _, err := os.Stat(cfg.WeightPath); os.IsNotExist(err) {
				fmt.Printf("[NeuroGo] Weight file '%s' not found. Auto-training...\n", cfg.WeightPath)
				if err := trainer.TrainFromConfig(cfg); err != nil {
					fmt.Printf("Auto-training failed: %v\n", err)
					os.Exit(1)
				}
			}
		}

		// 트랜스파일 파일 생성
		targetGo := strings.TrimSuffix(ngoFile, filepath.Ext(ngoFile)) + ".go"
		if err := os.WriteFile(targetGo, []byte(goCode), 0644); err != nil {
			fmt.Printf("Error writing generated Go file: %v\n", err)
			os.Exit(1)
		}

		if _, err := exec.LookPath("go"); err != nil {
			fmt.Printf("[NeuroGo] Transpiled to '%s'. ('go' binary not found in current PATH to execute directly)\n", targetGo)
			return
		}

		// 대상 파일이 위치한 디렉터리를 Working Directory로 설정하여 상대 경로 충돌 방지
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

		goCode, _, err := loadAndTranspile(ngoFile)
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
			fmt.Printf("[NeuroGo] Transpiled to '%s'. ('go' binary not found in current PATH to build directly)\n", targetGo)
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
