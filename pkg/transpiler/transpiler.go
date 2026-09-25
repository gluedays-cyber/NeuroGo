package transpiler

import (
	"fmt"
	"regexp"
	"strings"
)

// TrainConfig holds training configuration extracted from .ngo source files.
type TrainConfig struct {
	WeightPath string
	Source     string
	Input      string
	Target     string
	Epochs     int
	Format     string
}

// Regex patterns for NeuroGo syntax
var (
	// Matches: match <target> using "<modelPath>" {
	matchStartRegex = regexp.MustCompile(`^\s*match\s+(.+?)\s+using\s+"([^"]+)"\s*\{\s*$`)

	// Matches: case "<label>" score >= <identifier_or_number>:
	// [a-zA-Z0-9_\.] 패턴을 통해 숫자 리터럴(0.85)뿐만 아니라 변수명, 상수명(DefaultThreshold 등) 허용
	caseScoreRegex = regexp.MustCompile(`^\s*case\s+"([^"]+)"\s+score\s*>=\s*([a-zA-Z0-9_\.]+)\s*:\s*$`)

	// Matches: default:
	defaultRegex = regexp.MustCompile(`^\s*default\s*:\s*$`)
)

// Transpile converts .ngo source code into standard Go code, returning (goCode, referencedModels, error).
func Transpile(source string) (string, []string, error) {
	lines := strings.Split(source, "\n")
	var output []string
	var models []string

	inMatchBlock := false
	matchVarIndex := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 1. Detect 'match ... using ... {'
		if matchStartRegex.MatchString(trimmed) {
			if inMatchBlock {
				return "", nil, fmt.Errorf("line %d: nested match blocks are not supported", lineNum+1)
			}
			matches := matchStartRegex.FindStringSubmatch(trimmed)
			targetExpr := matches[1]
			modelPath := matches[2]

			models = append(models, modelPath)
			inMatchBlock = true
			varName := fmt.Sprintf("_ngoMatch%d", matchVarIndex)
			matchVarIndex++

			indent := getLeadingWhitespace(line)

			// Generate Go runtime match invocation and switch header
			output = append(output, fmt.Sprintf("%s{", indent))
			output = append(output, fmt.Sprintf("%s\t%s := runtime.Match(\"%s\", %s)", indent, varName, modelPath, targetExpr))
			output = append(output, fmt.Sprintf("%s\tswitch {", indent))
			continue
		}

		// 2. Detect 'case "<label>" score >= <identifier_or_number>:'
		if inMatchBlock && caseScoreRegex.MatchString(trimmed) {
			matches := caseScoreRegex.FindStringSubmatch(trimmed)
			label := matches[1]
			threshold := matches[2]

			indent := getLeadingWhitespace(line)
			varName := fmt.Sprintf("_ngoMatch%d", matchVarIndex-1)

			output = append(output, fmt.Sprintf("%scase %s.Label == \"%s\" && %s.Score >= %s:", indent, varName, label, varName, threshold))
			continue
		}

		// 3. Detect 'default:'
		if inMatchBlock && defaultRegex.MatchString(trimmed) {
			output = append(output, line)
			continue
		}

		// 4. Detect closing brace '}' for match block
		if inMatchBlock && trimmed == "}" {
			inMatchBlock = false
			indent := getLeadingWhitespace(line)

			// Close both the switch and the outer block scope
			output = append(output, fmt.Sprintf("%s}", indent))
			output = append(output, fmt.Sprintf("%s}", indent))
			continue
		}

		// Standard Go lines pass through unchanged
		output = append(output, line)
	}

	if inMatchBlock {
		return "", nil, fmt.Errorf("unexpected EOF: unclosed match block")
	}

	result := strings.Join(output, "\n")

	// Ensure runtime package import exists if runtime.Match is invoked
	if matchVarIndex > 0 && !strings.Contains(result, `"neurogo/pkg/runtime"`) && !strings.Contains(result, `"NeuroGo/pkg/runtime"`) && !strings.Contains(result, `"runtime"`) {
		result = injectRuntimeImport(result)
	}

	return result, models, nil
}

func getLeadingWhitespace(s string) string {
	var ws []rune
	for _, r := range s {
		if r == ' ' || r == '\t' {
			ws = append(ws, r)
		} else {
			break
		}
	}
	return string(ws)
}

func injectRuntimeImport(code string) string {
	importPattern := regexp.MustCompile(`import\s*\(([\s\S]*?)\)`)
	if importPattern.MatchString(code) {
		return importPattern.ReplaceAllString(code, "import (\n\t\"neurogo/pkg/runtime\"$1)")
	}

	singleImport := regexp.MustCompile(`import\s+"([^"]+)"`)
	if singleImport.MatchString(code) {
		return singleImport.ReplaceAllString(code, "import (\n\t\"neurogo/pkg/runtime\"\n\t\"$1\"\n)")
	}

	packagePattern := regexp.MustCompile(`package\s+[a-zA-Z0-9_]+`)
	return packagePattern.ReplaceAllString(code, "$0\n\nimport \"neurogo/pkg/runtime\"")
}
