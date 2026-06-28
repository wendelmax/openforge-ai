//go:build cgo

package openvino

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/openforge-ai/openforge/runtime"
)

func (r *OpenVINORuntime) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
	r.mu.RLock()
	lm, ok := r.models[modelID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("model %q is not loaded", modelID)
	}

	inputIDs, err := r.tokenizer.Encode(prompt)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}

	if params.MaxTokens <= 0 {
		params.MaxTokens = 2048
	}

	outputIDs, err := r.generate(ctx, lm, inputIDs, params)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	text, err := r.tokenizer.Decode(outputIDs)
	if err != nil {
		return nil, fmt.Errorf("detokenization failed: %w", err)
	}

	tokenInts := make([]int, len(outputIDs))
	for i, id := range outputIDs {
		tokenInts[i] = int(id)
	}
	return &runtime.InferenceResult{
		Text:   text,
		Tokens: tokenInts,
	}, nil
}

func (r *OpenVINORuntime) InferStream(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (<-chan string, error) {
	r.mu.RLock()
	lm, ok := r.models[modelID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("model %q is not loaded", modelID)
	}

	device := params.Device
	if device == "" || device == "auto" {
		device = r.defaultDev
	}

	ch, ok := lm.compiledByDevice[device]
	if !ok {
		return nil, fmt.Errorf("model %q not compiled for device %q", modelID, device)
	}

	inputIDs, err := r.tokenizer.Encode(prompt)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}

	if params.MaxTokens <= 0 {
		params.MaxTokens = 2048
	}

	out := make(chan string, 8)

	go func() {
		defer close(out)

		req, err := ch.compiled.CreateInferRequest()
		if err != nil {
			return
		}
		defer req.Free()

		ids := make([]int64, len(inputIDs))
		copy(ids, inputIDs)

		for i := 0; i < params.MaxTokens; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var nextID int64
			if i == 0 {
				logits, err := r.runInferenceFull(req, ids)
				if err != nil {
					return
				}
				nextID = sampleToken(logits, r.tokenizer.VocabSize(), params)
			} else {
				logits, err := r.runInferenceIncremental(req, ids[len(ids)-1:])
				if err != nil {
					return
				}
				nextID = sampleToken(logits, r.tokenizer.VocabSize(), params)
			}

			if nextID < 0 {
				break
			}

			ids = append(ids, nextID)

			tokenText, _ := r.tokenizer.Decode([]int64{nextID})
			if tokenText != "" {
				select {
				case out <- tokenText:
				case <-ctx.Done():
					return
				}
			}

			if isEOS(nextID) {
				break
			}
		}
	}()

	return out, nil
}

func (r *OpenVINORuntime) generate(ctx context.Context, lm *loadedModel, inputIDs []int64, params runtime.InferenceParams) ([]int64, error) {
	device := params.Device
	if device == "" || device == "auto" {
		device = r.defaultDev
	}
	ch, ok := lm.compiledByDevice[device]
	if !ok {
		return nil, fmt.Errorf("model not compiled for device %q", device)
	}

	ids := make([]int64, len(inputIDs))
	copy(ids, inputIDs)
	outputIDs := make([]int64, 0, params.MaxTokens)
	vocabSize := r.tokenizer.VocabSize()

	req, err := ch.compiled.CreateInferRequest()
	if err != nil {
		return nil, fmt.Errorf("create infer request: %w", err)
	}
	defer req.Free()

	for i := 0; i < params.MaxTokens; i++ {
		select {
		case <-ctx.Done():
			return outputIDs, ctx.Err()
		default:
		}

		var logits []float32
		if i == 0 {
			logits, err = r.runInferenceFull(req, ids)
		} else {
			logits, err = r.runInferenceIncremental(req, ids[len(ids)-1:])
		}
		if err != nil {
			return outputIDs, fmt.Errorf("inference step %d failed: %w", i, err)
		}

		nextID := sampleToken(logits, vocabSize, params)
		if nextID < 0 {
			break
		}

		outputIDs = append(outputIDs, nextID)
		ids = append(ids, nextID)

		if isEOS(nextID) {
			break
		}
	}

	return outputIDs, nil
}

func (r *OpenVINORuntime) nextToken(ctx context.Context, lm *loadedModel, compiled *CompiledModel, ids []int64, params runtime.InferenceParams) (int64, error) {
	logits, err := r.runInference(compiled, ids)
	if err != nil {
		return -1, fmt.Errorf("next token inference failed: %w", err)
	}

	nextID := sampleToken(logits, r.tokenizer.VocabSize(), params)
	return nextID, nil
}

func (r *OpenVINORuntime) runInference(compiled *CompiledModel, inputIDs []int64) ([]float32, error) {
	req, err := compiled.CreateInferRequest()
	if err != nil {
		return nil, err
	}
	defer req.Free()

	return r.inferWithRequest(req, inputIDs)
}

func (r *OpenVINORuntime) runInferenceFull(req *InferRequest, inputIDs []int64) ([]float32, error) {
	return r.inferWithRequest(req, inputIDs)
}

func (r *OpenVINORuntime) runInferenceIncremental(req *InferRequest, inputIDs []int64) ([]float32, error) {
	return r.inferWithRequest(req, inputIDs)
}

func (r *OpenVINORuntime) inferWithRequest(req *InferRequest, inputIDs []int64) ([]float32, error) {
	seqLen := int64(len(inputIDs))

	tensor, err := NewTensor(ElementTypeI64, []int64{1, seqLen})
	if err != nil {
		return nil, fmt.Errorf("create input tensor: %w", err)
	}
	defer tensor.Free()

	data := (*[1 << 30]int64)(tensor.Data())[:seqLen]
	for i, id := range inputIDs {
		data[i] = id
	}

	if err := req.SetInputTensor(0, tensor); err != nil {
		return nil, err
	}

	if err := req.Infer(); err != nil {
		return nil, err
	}

	outputTensor, err := req.GetOutputTensor(0)
	if err != nil {
		return nil, err
	}
	defer outputTensor.Free()

	outData := (*[1 << 30]float32)(outputTensor.Data())
	vocabSize := int64(r.tokenizer.VocabSize())
	logits := outData[(seqLen-1)*vocabSize : seqLen*vocabSize]
	result := make([]float32, len(logits))
	copy(result, logits)

	return result, nil
}

func sampleToken(logits []float32, vocabSize int, params runtime.InferenceParams) int64 {
	if len(logits) == 0 {
		return -1
	}

	temp := params.Temperature
	if temp <= 0 {
		temp = 0.1
	}

	topK := params.TopK
	if topK <= 0 {
		topK = 40
	}

	topP := params.TopP
	if topP <= 0 {
		topP = 0.9
	}

	scaled := make([]float32, len(logits))
	for i, l := range logits {
		scaled[i] = l / temp
	}

	maxLogit := float32(-1e9)
	for _, l := range scaled {
		if l > maxLogit {
			maxLogit = l
		}
	}

	sum := float32(0)
	probs := make([]float32, len(scaled))
	for i, l := range scaled {
		probs[i] = exp(l - maxLogit)
		sum += probs[i]
	}

	if sum > 0 {
		for i := range probs {
			probs[i] /= sum
		}
	}

	tokens := make([]tokenProb, len(probs))
	for i, p := range probs {
		tokens[i] = tokenProb{id: i, prob: p}
	}

	sortByProb(tokens)

	if topK > 0 && topK < len(tokens) {
		tokens = tokens[:topK]
	}

	var cumsum float32
	cutoff := 0
	for i, t := range tokens {
		cumsum += t.prob
		if cumsum >= topP {
			cutoff = i + 1
			break
		}
	}
	if cutoff > 0 && cutoff < len(tokens) {
		tokens = tokens[:cutoff]
	}

	rnd := rand.Float32()
	var tprobsum float32
	for _, t := range tokens {
		tprobsum += t.prob
	}
	if tprobsum > 0 {
		rnd *= tprobsum
		cumsum = 0
		for _, t := range tokens {
			cumsum += t.prob
			if rnd <= cumsum {
				return int64(t.id)
			}
		}
	}

	if len(tokens) > 0 {
		return int64(tokens[0].id)
	}

	return 0
}

func serializeEmbeddings(embeddings [][]float32) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int32(len(embeddings)))
	for _, emb := range embeddings {
		binary.Write(&buf, binary.LittleEndian, int32(len(emb)))
		for _, v := range emb {
			binary.Write(&buf, binary.LittleEndian, v)
		}
	}
	return buf.Bytes()
}

func deserializeEmbeddings(data []byte) ([][]float32, error) {
	buf := bytes.NewReader(data)
	var numEmbs int32
	if err := binary.Read(buf, binary.LittleEndian, &numEmbs); err != nil {
		return nil, err
	}
	embeddings := make([][]float32, numEmbs)
	for i := int32(0); i < numEmbs; i++ {
		var dim int32
		if err := binary.Read(buf, binary.LittleEndian, &dim); err != nil {
			return nil, err
		}
		emb := make([]float32, dim)
		for j := int32(0); j < dim; j++ {
			if err := binary.Read(buf, binary.LittleEndian, &emb[j]); err != nil {
				return nil, err
			}
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (r *OpenVINORuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	r.mu.RLock()
	lm, ok := r.models[modelID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("model %q is not loaded", modelID)
	}

	resolvedDevice := device
	if resolvedDevice == "" || resolvedDevice == "auto" {
		resolvedDevice = r.defaultDev
	}
	if resolvedDevice == "" {
		resolvedDevice = "CPU"
	}

	embeddings := make([][]float32, len(inputs))

	for i, input := range inputs {
		cacheKey := modelID + ":" + input
		if cached, found := r.embedCache.Get(ctx, cacheKey); found {
			if embs, err := deserializeEmbeddings(cached); err == nil && len(embs) > 0 {
				embeddings[i] = embs[0]
				continue
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		inputIDs, err := r.tokenizer.Encode(input)
		if err != nil {
			return nil, fmt.Errorf("tokenization failed for input %d: %w", i, err)
		}

		emb, err := r.extractEmbedding(lm, resolvedDevice, inputIDs)
		if err != nil {
			return nil, fmt.Errorf("embedding failed for input %d: %w", i, err)
		}

		embeddings[i] = emb

		singleEmb := [][]float32{emb}
		r.embedCache.Set(ctx, cacheKey, serializeEmbeddings(singleEmb), 5*time.Minute)
	}

	return &runtime.EmbeddingResult{
		Embeddings: embeddings,
	}, nil
}

func (r *OpenVINORuntime) extractEmbedding(lm *loadedModel, device string, inputIDs []int64) ([]float32, error) {
	ch, ok := lm.compiledByDevice[device]
	if !ok {
		return nil, fmt.Errorf("model not compiled for device %q", device)
	}
	req, err := ch.compiled.CreateInferRequest()
	if err != nil {
		return nil, err
	}
	defer req.Free()

	batchSize := int64(1)
	seqLen := int64(len(inputIDs))
	shape := []int64{batchSize, seqLen}

	tensor, err := NewTensor(	ElementTypeI64, shape)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer tensor.Free()

	data := (*[1 << 30]int64)(tensor.Data())[:seqLen]
	for i, id := range inputIDs {
		data[i] = id
	}

	if err := req.SetInputTensor(0, tensor); err != nil {
		return nil, err
	}

	if err := req.Infer(); err != nil {
		return nil, err
	}

	outputCount, err := lm.model.OutputsCount()
	if err != nil {
		return nil, err
	}

	outputIdx := 0
	if outputCount > 1 {
		outputIdx = 1
	}

	outputTensor, err := req.GetOutputTensor(outputIdx)
	if err != nil {
		return nil, err
	}
	defer outputTensor.Free()

	outData := (*[1 << 30]float32)(outputTensor.Data())

	outSeqLen := seqLen
	outDim := int64(len(outData)) / (batchSize * outSeqLen)
	if outDim == 0 {
		outDim = int64(len(outData)) / batchSize
		outSeqLen = 1
	}

	embedDim := int64(768)
	if outDim > 0 {
		embedDim = outDim
	}

	pooled := meanPool(outData[:batchSize*outSeqLen*embedDim], int(outSeqLen), int(embedDim))
	return pooled, nil
}

func meanPool(data []float32, seqLen, dim int) []float32 {
	result := make([]float32, dim)
	if seqLen == 0 {
		return result
	}
	for i := 0; i < seqLen; i++ {
		for j := 0; j < dim; j++ {
			result[j] += data[i*dim+j]
		}
	}
	for j := 0; j < dim; j++ {
		result[j] /= float32(seqLen)
	}
	return result
}

var eosTokens = map[int64]bool{
	2: true,
}

func isEOS(id int64) bool {
	return eosTokens[id]
}

var (
	rngPool = sync.Pool{
		New: func() interface{} {
			return rand.New(rand.NewSource(time.Now().UnixNano()))
		},
	}
)

func exp(x float32) float32 {
	const limit = 80.0
	if x > limit {
		x = limit
	}
	if x < -limit {
		x = -limit
	}

	result := float32(1.0)
	term := float32(1.0)
	for i := 1; i <= 30; i++ {
		term *= x / float32(i)
		result += term
	}
	return result
}

func sortByProb(tokens []tokenProb) {
	n := len(tokens)
	for i := 0; i < n-1; i++ {
		maxIdx := i
		for j := i + 1; j < n; j++ {
			if tokens[j].prob > tokens[maxIdx].prob {
				maxIdx = j
			}
		}
		if maxIdx != i {
			tokens[i], tokens[maxIdx] = tokens[maxIdx], tokens[i]
		}
	}
}

type tokenProb struct {
	id   int
	prob float32
}
