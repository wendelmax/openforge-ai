package openvino

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteLevelTokenizer_EncodeDecode(t *testing.T) {
	tok := NewByteLevelTokenizer(50272)

	ids, err := tok.Encode("hello")
	assert.NoError(t, err)
	assert.NotEmpty(t, ids)

	text, err := tok.Decode(ids)
	assert.NoError(t, err)
	assert.Equal(t, "hello", text)
}

func TestByteLevelTokenizer_Empty(t *testing.T) {
	tok := NewByteLevelTokenizer(50272)

	ids, err := tok.Encode("")
	assert.NoError(t, err)
	assert.Empty(t, ids)
}

func TestByteLevelTokenizer_VocabSize(t *testing.T) {
	tok := NewByteLevelTokenizer(50272)
	assert.Equal(t, 50272, tok.VocabSize())
}

func TestWhitespaceTokenizer_EncodeDecode(t *testing.T) {
	tok := NewWhitespaceTokenizer(50000)

	ids, err := tok.Encode("hello world")
	assert.NoError(t, err)
	assert.Len(t, ids, 2)

	text, err := tok.Decode(ids)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", text)
}

func TestWhitespaceTokenizer_ConsistentTokens(t *testing.T) {
	tok := NewWhitespaceTokenizer(50000)

	ids1, _ := tok.Encode("hello hello")
	ids2, _ := tok.Encode("hello")
	assert.Equal(t, ids1[0], ids2[0])
	assert.Equal(t, ids1[1], ids2[0])
}

func TestByteLevelTokenizer_UTF8(t *testing.T) {
	tok := NewByteLevelTokenizer(50272)

	ids, err := tok.Encode("olá mundo")
	assert.NoError(t, err)
	assert.NotEmpty(t, ids)

	text, err := tok.Decode(ids)
	assert.NoError(t, err)
	assert.Equal(t, "olá mundo", text)
}
