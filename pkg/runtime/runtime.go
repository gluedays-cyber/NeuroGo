package runtime

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"unicode"
)

var MagicGOW = [4]byte{'G', 'O', 'W', 1}

type Header struct {
	Magic        [4]byte        `json:"magic"`
	Version      int            `json:"version"`
	Architecture string         `json:"architecture"`
	InputDim     int            `json:"input_dim"`
	NumClasses   int            `json:"num_classes"`
	Labels       []string       `json:"labels"`
	Vocab        map[string]int `json:"vocab"`
}

type Model struct {
	Header  Header
	Weights []float32 // flattened shape: [InputDim, NumClasses]
	Biases  []float32 // shape: [NumClasses]
}

type MatchResult struct {
	Label  string             `json:"label"`
	Score  float64            `json:"score"`
	Scores map[string]float64 `json:"scores"`
}

var (
	modelCache = make(map[string]*Model)
	cacheMu    sync.RWMutex
)

func LoadModel(path string) (*Model, error) {
	cacheMu.RLock()
	if m, ok := modelCache[path]; ok {
		cacheMu.RUnlock()
		return m, nil
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if m, ok := modelCache[path]; ok {
		return m, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read .gow file %s: %w", path, err)
	}

	if len(data) < 8 {
		return nil, fmt.Errorf("invalid .gow file: file too short")
	}

	if !bytes.Equal(data[:4], MagicGOW[:]) {
		return nil, fmt.Errorf("invalid .gow magic header")
	}

	headerLen := binary.LittleEndian.Uint32(data[4:8])
	if int(8+headerLen) > len(data) {
		return nil, fmt.Errorf("corrupted .gow header length")
	}

	var header Header
	if err := json.Unmarshal(data[8:8+headerLen], &header); err != nil {
		return nil, fmt.Errorf("failed to parse .gow header JSON: %w", err)
	}

	offset := 8 + headerLen
	expectedFloats := header.InputDim*header.NumClasses + header.NumClasses
	expectedBytes := expectedFloats * 4

	if int(offset)+expectedBytes > len(data) {
		return nil, fmt.Errorf("insufficient tensor data in .gow file")
	}

	weights := make([]float32, header.InputDim*header.NumClasses)
	biases := make([]float32, header.NumClasses)

	r := bytes.NewReader(data[offset:])
	if err := binary.Read(r, binary.LittleEndian, &weights); err != nil {
		return nil, fmt.Errorf("failed to read weights: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &biases); err != nil {
		return nil, fmt.Errorf("failed to read biases: %w", err)
	}

	model := &Model{
		Header:  header,
		Weights: weights,
		Biases:  biases,
	}

	modelCache[path] = model
	return model, nil
}

// Tokenize splits text into words and 2-gram subwords for robust morphology handling.
func Tokenize(text string) []string {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	rawWords := strings.FieldsFunc(text, f)
	tokens := make([]string, 0, len(rawWords)*3)

	for _, w := range rawWords {
		lower := strings.ToLower(w)
		runes := []rune(lower)
		if len(runes) == 0 {
			continue
		}
		tokens = append(tokens, lower)

		// Character bi-grams
		if len(runes) >= 2 {
			for i := 0; i <= len(runes)-2; i++ {
				tokens = append(tokens, string(runes[i:i+2]))
			}
		}
	}
	return tokens
}

func (m *Model) Vectorize(input any) ([]float32, error) {
	vec := make([]float32, m.Header.InputDim)

	switch val := input.(type) {
	case string:
		tokens := Tokenize(val)
		for _, token := range tokens {
			if idx, ok := m.Header.Vocab[token]; ok && idx < m.Header.InputDim {
				vec[idx] += 1.0
			}
		}
	case []float64:
		for i := 0; i < len(val) && i < m.Header.InputDim; i++ {
			vec[i] = float32(val[i])
		}
	case []float32:
		copy(vec, val)
	default:
		return nil, fmt.Errorf("unsupported input type for vectorization: %T", input)
	}

	return vec, nil
}

func (m *Model) Forward(vec []float32) MatchResult {
	numClasses := m.Header.NumClasses
	inputDim := m.Header.InputDim
	logits := make([]float64, numClasses)

	for c := 0; c < numClasses; c++ {
		sum := float64(m.Biases[c])
		for i := 0; i < inputDim; i++ {
			if vec[i] != 0 {
				sum += float64(vec[i]) * float64(m.Weights[i*numClasses+c])
			}
		}
		logits[c] = sum
	}

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

	scores := make(map[string]float64, numClasses)
	bestLabel := ""
	bestScore := -1.0

	for c := 0; c < numClasses; c++ {
		score := probs[c] / expSum
		label := m.Header.Labels[c]
		scores[label] = score
		if score > bestScore {
			bestScore = score
			bestLabel = label
		}
	}

	return MatchResult{
		Label:  bestLabel,
		Score:  bestScore,
		Scores: scores,
	}
}

func Match(weightPath string, input any) MatchResult {
	m, err := LoadModel(weightPath)
	if err != nil {
		return MatchResult{
			Label:  "Error",
			Score:  0.0,
			Scores: map[string]float64{"error": 0.0},
		}
	}

	vec, err := m.Vectorize(input)
	if err != nil {
		return MatchResult{
			Label:  "Error",
			Score:  0.0,
			Scores: map[string]float64{"error": 0.0},
		}
	}

	return m.Forward(vec)
}

func SaveGOW(path string, header Header, weights []float32, biases []float32) error {
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(MagicGOW[:])
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(headerBytes))); err != nil {
		return err
	}
	buf.Write(headerBytes)

	if err := binary.Write(buf, binary.LittleEndian, weights); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, biases); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

var _ io.Reader
