package openvino

import "strings"

type Tokenizer interface {
	Encode(text string) ([]int64, error)
	Decode(tokens []int64) (string, error)
	VocabSize() int
}

type ByteLevelTokenizer struct {
	vocabSize int
	padToken  int64
	unkToken  int64
	eosToken  int64
	boost     map[string]int64
}

func NewByteLevelTokenizer(vocabSize int) *ByteLevelTokenizer {
	return &ByteLevelTokenizer{
		vocabSize: vocabSize,
		padToken:  0,
		unkToken:  1,
		eosToken:  2,
		boost:     make(map[string]int64),
	}
}

func (t *ByteLevelTokenizer) Encode(text string) ([]int64, error) {
	ids := make([]int64, 0, len(text))
	for _, b := range []byte(text) {
		id := int64(b) % int64(t.vocabSize)
		ids = append(ids, id)
	}
	return ids, nil
}

func (t *ByteLevelTokenizer) Decode(tokens []int64) (string, error) {
	var sb strings.Builder
	for _, id := range tokens {
		if id < 256 {
			sb.WriteByte(byte(id))
		}
	}
	return sb.String(), nil
}

func (t *ByteLevelTokenizer) VocabSize() int {
	return t.vocabSize
}

type WhitespaceTokenizer struct {
	vocabSize int
	words     map[string]int64
	ids       map[int64]string
}

func NewWhitespaceTokenizer(vocabSize int) *WhitespaceTokenizer {
	return &WhitespaceTokenizer{
		vocabSize: vocabSize,
		words:     make(map[string]int64),
		ids:       make(map[int64]string),
	}
}

func (t *WhitespaceTokenizer) Encode(text string) ([]int64, error) {
	parts := strings.Fields(text)
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		if id, ok := t.words[part]; ok {
			ids = append(ids, id)
		} else {
			id := int64(len(t.words))
			if id < int64(t.vocabSize) {
				t.words[part] = id
				t.ids[id] = part
				ids = append(ids, id)
			}
		}
	}
	return ids, nil
}

func (t *WhitespaceTokenizer) Decode(tokens []int64) (string, error) {
	var sb strings.Builder
	for i, id := range tokens {
		if word, ok := t.ids[id]; ok {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(word)
		}
	}
	return sb.String(), nil
}

func (t *WhitespaceTokenizer) VocabSize() int {
	return t.vocabSize
}

type CompositeTokenizer struct {
	tokenizers []Tokenizer
}

func (c *CompositeTokenizer) Encode(text string) ([]int64, error) {
	var ids []int64
	for _, t := range c.tokenizers {
		part, err := t.Encode(text)
		if err != nil {
			return nil, err
		}
		ids = append(ids, part...)
	}
	return ids, nil
}

func (c *CompositeTokenizer) Decode(tokens []int64) (string, error) {
	var sb strings.Builder
	for _, t := range c.tokenizers {
		part, err := t.Decode(tokens)
		if err != nil {
			return "", err
		}
		sb.WriteString(part)
	}
	return sb.String(), nil
}

func (c *CompositeTokenizer) VocabSize() int {
	if len(c.tokenizers) == 0 {
		return 0
	}
	return c.tokenizers[0].VocabSize()
}
