package transpiler

import (
	"fmt"
	"regexp"
	"strings"
)

// TrainConfig holds the parsed metadata from a `train` block in .ngo source.
type TrainConfig struct {
	WeightPath string
	Format     string
	Source     string
	InputCol   string
	TargetCol  string
	Epochs     int
}

// Transpile converts a NeuroGo (.ngo) source code string into valid standard Go (.go) code.
func Transpile(source string) (string, []TrainConfig, error) {
	trainConfigs, strippedSource, err := extractTrainBlocks(source)
	if err != nil {
		return "", nil, err
	}

	transformedSource, err := transformMatchBlocks(strippedSource)
	if err != nil {
		return "", nil, err
	}

	finalSource := ensureRuntimeImport(transformedSource)
	return finalSource, trainConfigs, nil
}

// extractTrainBlocks extracts `train` declarations and replaces them with comments.
var trainRegex = regexp.MustCompile(`(?s)train\s+"([^"]+)"\s*\{([^}]+)\}`)

func extractTrainBlocks(src string) ([]TrainConfig, string, error) {
	var configs []TrainConfig
	matches := trainRegex.FindAllStringSubmatchIndex(src, -1)
	if len(matches) == 0 {
		return configs, src, nil
	}

	var sb strings.Builder
	lastIdx := 0

	for _, m := range matches {
		sb.WriteString(src[lastIdx:m[0]])
		weightPath := src[m[2]:m[3]]
		body := src[m[4]:m[5]]

		cfg := TrainConfig{
			WeightPath: weightPath,
			Epochs:     30,
			Format:     "csv",
			InputCol:   "text",
			TargetCol:  "label",
		}

		// Simple key-value parser for train body
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "format:") {
				cfg.Format = extractStringVal(line)
			} else if strings.HasPrefix(line, "source:") {
				cfg.Source = extractStringVal(line)
			} else if strings.HasPrefix(line, "input:") {
				cfg.InputCol = extractStringVal(line)
			} else if strings.HasPrefix(line, "target:") {
				cfg.TargetCol = extractStringVal(line)
			} else if strings.HasPrefix(line, "epochs:") {
				var ep int
				fmt.Sscanf(line, "epochs: %d", &ep)
				if ep > 0 {
					cfg.Epochs = ep
				}
			}
		}

		configs = append(configs, cfg)
		sb.WriteString(fmt.Sprintf("/* [NeuroGo] train \"%s\" extracted for pre-compilation */", weightPath))
		lastIdx = m[1]
	}

	sb.WriteString(src[lastIdx:])
	return configs, sb.String(), nil
}

func extractStringVal(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	val := strings.TrimSpace(parts[1])
	val = strings.Trim(val, `",; `)
	return val
}

// transformMatchBlocks converts `match <expr> using <weight> { ... }` into Go switch statements.
var matchHeaderRegex = regexp.MustCompile(`match\s+(.+?)\s+using\s+"([^"]+)"\s*\{`)
var caseScoreRegex = regexp.MustCompile(`case\s+"([^"]+)"\s+score\s*(>=|>|<=|<|==)\s*([0-9.]+)\s*:`)

func transformMatchBlocks(src string) (string, error) {
	// Find all match blocks
	out := src
	for {
		loc := matchHeaderRegex.FindStringSubmatchIndex(out)
		if loc == nil {
			break
		}

		headerStart := loc[0]
		bodyStart := loc[1] // right after '{'

		expr := strings.TrimSpace(out[loc[2]:loc[3]])
		weightPath := out[loc[4]:loc[5]]

		// Find matching closing brace '}'
		braceCount := 1
		endIdx := -1
		for i := bodyStart; i < len(out); i++ {
			if out[i] == '{' {
				braceCount++
			} else if out[i] == '}' {
				braceCount--
				if braceCount == 0 {
					endIdx = i
					break
				}
			}
		}

		if endIdx == -1 {
			return "", fmt.Errorf("unmatched brace in match block starting at offset %d", headerStart)
		}

		bodyContent := out[bodyStart:endIdx]

		// Replace case clauses inside the body
		transformedBody := caseScoreRegex.ReplaceAllStringFunc(bodyContent, func(caseClause string) string {
			m := caseScoreRegex.FindStringSubmatch(caseClause)
			if len(m) < 4 {
				return caseClause
			}
			label := m[1]
			op := m[2]
			threshold := m[3]
			return fmt.Sprintf("case _ngoMatch.Label == \"%s\" && _ngoMatch.Score %s %s:", label, op, threshold)
		})

		replacement := fmt.Sprintf("{\n\t_ngoMatch := runtime.Match(\"%s\", %s)\n\tswitch {\n%s\n\t}\n}",
			weightPath, expr, transformedBody)

		out = out[:headerStart] + replacement + out[endIdx+1:]
	}

	return out, nil
}

// ensureRuntimeImport makes sure "neurogo/pkg/runtime" is imported in the generated Go code.
func ensureRuntimeImport(src string) string {
	if strings.Contains(src, `"neurogo/pkg/runtime"`) {
		return src
	}

	pkgIdx := strings.Index(src, "package ")
	if pkgIdx == -1 {
		return src
	}

	lineEnd := strings.Index(src[pkgIdx:], "\n")
	if lineEnd == -1 {
		return src
	}
	insertPos := pkgIdx + lineEnd + 1

	importStmt := "\nimport \"neurogo/pkg/runtime\"\n"
	return src[:insertPos] + importStmt + src[insertPos:]
}
