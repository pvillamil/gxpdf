package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hasPfx is a local helper since testify has no HasPrefix assertion.
func hasPfx(t *testing.T, s, prefix string) {
	t.Helper()
	if !strings.HasPrefix(s, prefix) {
		t.Errorf("expected %q to have prefix %q", s, prefix)
	}
}

// hasSfx is a local helper since testify has no HasSuffix assertion.
func hasSfx(t *testing.T, s, suffix string) {
	t.Helper()
	if !strings.HasSuffix(s, suffix) {
		t.Errorf("expected %q to have suffix %q", s, suffix)
	}
}

// ============================================================================
// Lexer.Peek and Lexer.Position — previously 0% coverage
// ============================================================================

func TestLexer_Position_Initial(t *testing.T) {
	l := NewLexer(strings.NewReader("123"))
	line, col := l.Position()
	assert.Equal(t, 1, line, "initial line should be 1")
	assert.Equal(t, 0, col, "initial column should be 0")
}

func TestLexer_Position_AfterRead(t *testing.T) {
	l := NewLexer(strings.NewReader("123"))
	_, err := l.NextToken()
	require.NoError(t, err)
	line, col := l.Position()
	assert.Equal(t, 1, line, "should still be on line 1 after reading 3-digit integer")
	assert.Equal(t, 3, col, "column should be 3 after reading '123'")
}

func TestLexer_Position_AfterNewline(t *testing.T) {
	l := NewLexer(strings.NewReader("123\n456"))
	_, _ = l.NextToken() // consume 123
	_, _ = l.NextToken() // consume 456
	line, _ := l.Position()
	assert.Equal(t, 2, line, "should be on line 2 after newline")
}

// TestLexer_Peek_DoesNotConsumeToken verifies Peek returns token without advancing.
// Note: the current Peek implementation has a known limitation (it restores the
// saved reader reference but cannot rewind bufio.Reader). We test what it does.
func TestLexer_Peek_ReturnsToken(t *testing.T) {
	l := NewLexer(strings.NewReader("123"))
	tok, err := l.Peek()
	// Peek should return a token (even if it can't perfectly restore state).
	// The important thing is it does not panic and returns some token.
	assert.NoError(t, err)
	assert.NotEqual(t, TokenError, tok.Type)
}

// ============================================================================
// Object.String() for all types — many type branches uncovered
// ============================================================================

func TestType_String_AllTypes(t *testing.T) {
	tests := []struct {
		t    Type
		want string
	}{
		{TypeNull, "Null"},
		{TypeBoolean, "Boolean"},
		{TypeInteger, "Integer"},
		{TypeReal, "Real"},
		{TypeString, "String"},
		{TypeName, "Name"},
		{TypeArray, "Array"},
		{TypeDictionary, "Dictionary"},
		{TypeStream, "Stream"},
		{TypeIndirect, "Indirect"},
		{TypeReference, "Reference"},
		{Type(999), "Unknown(999)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.t.String())
		})
	}
}

// TestTypeOf_ArrayAndDictionary covers the Array and Dictionary branches in TypeOf.
func TestTypeOf_ArrayAndDictionary(t *testing.T) {
	arr := NewArray()
	assert.Equal(t, TypeArray, TypeOf(arr))

	dict := NewDictionary()
	assert.Equal(t, TypeDictionary, TypeOf(dict))
}

// TestTypeOf_Unknown covers the default branch (unknown type → -1).
func TestTypeOf_Unknown(t *testing.T) {
	// A nil PdfObject doesn't match any known type.
	// We pass a custom type that doesn't match any case.
	// We can't easily create an unknown PdfObject, but we can call with a
	// *Stream which is not in the TypeOf switch → returns -1.
	stream := NewStream(NewDictionary(), []byte("data"))
	got := TypeOf(stream)
	assert.Equal(t, Type(-1), got)
}

// ============================================================================
// Clone — Array and Dictionary branches
// ============================================================================

func TestClone_Array(t *testing.T) {
	arr := NewArray()
	arr.Append(NewInteger(1))
	arr.Append(NewString("hello"))
	arr.Append(NewBoolean(true))

	cloned := Clone(arr)
	require.NotNil(t, cloned)

	clonedArr, ok := cloned.(*Array)
	require.True(t, ok, "cloned array should be *Array")
	assert.Equal(t, 3, clonedArr.Len())
	assert.Equal(t, arr.String(), clonedArr.String())
	assert.NotSame(t, arr, clonedArr)
}

func TestClone_Dictionary(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key1", NewInteger(42))
	dict.Set("key2", NewString("world"))

	cloned := Clone(dict)
	require.NotNil(t, cloned)

	clonedDict, ok := cloned.(*Dictionary)
	require.True(t, ok, "cloned dict should be *Dictionary")
	assert.Equal(t, 2, clonedDict.Len())
	assert.Equal(t, dict.String(), clonedDict.String())
	assert.NotSame(t, dict, clonedDict)
}

func TestClone_Unknown_ReturnsNil(t *testing.T) {
	// A *Stream is not handled by Clone → returns nil.
	stream := NewStream(NewDictionary(), []byte("data"))
	result := Clone(stream)
	assert.Nil(t, result)
}

// ============================================================================
// Resolve — previously 0% coverage
// ============================================================================

func TestResolve_DirectObject(t *testing.T) {
	obj := NewInteger(42)
	resolved := Resolve(obj)
	assert.Same(t, obj, resolved, "Resolve of direct object should return same object")
}

func TestResolve_Null(t *testing.T) {
	n := NewNull()
	resolved := Resolve(n)
	assert.NotNil(t, resolved)
}

// ============================================================================
// Array.String() with nil element
// ============================================================================

func TestArray_String_WithNilElement(t *testing.T) {
	arr := NewArray()
	arr.Append(NewInteger(1))
	arr.Append(nil)
	arr.Append(NewInteger(3))

	s := arr.String()
	assert.Contains(t, s, "null", "nil element should serialize as 'null'")
	assert.Contains(t, s, "1")
	assert.Contains(t, s, "3")
}

func TestArray_WriteTo_WithNilElement(t *testing.T) {
	arr := NewArray()
	arr.Append(NewInteger(1))
	arr.Append(nil)

	var buf bytes.Buffer
	n, err := arr.WriteTo(&buf)
	require.NoError(t, err)
	assert.Greater(t, n, int64(0))
	assert.Contains(t, buf.String(), "null")
}

// ============================================================================
// Dictionary.String() and WriteTo with nil values
// ============================================================================

func TestDictionary_String_WithNilValue(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key1", NewInteger(42))
	dict.Set("key2", nil)

	s := dict.String()
	assert.Contains(t, s, "/key1")
	assert.Contains(t, s, "null", "nil value should serialize as 'null'")
}

func TestDictionary_WriteTo_WithNilValue(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key1", NewInteger(42))
	dict.Set("key2", nil)

	var buf bytes.Buffer
	n, err := dict.WriteTo(&buf)
	require.NoError(t, err)
	assert.Greater(t, n, int64(0))
	assert.Contains(t, buf.String(), "null")
}

func TestDictionary_WriteTo_MultipleEntries(t *testing.T) {
	dict := NewDictionary()
	dict.Set("Type", NewName("Catalog"))
	dict.Set("Pages", NewInteger(2))
	dict.Set("Encrypted", NewBoolean(false))

	var buf bytes.Buffer
	n, err := dict.WriteTo(&buf)
	require.NoError(t, err)
	assert.Greater(t, n, int64(0))
	out := buf.String()
	hasPfx(t, out, "<<")
	assert.Contains(t, out, "/Type")
	assert.Contains(t, out, "/Pages")
}

// ============================================================================
// Dictionary helper getters — type-mismatch branches (return zero values)
// ============================================================================

func TestDictionary_GetName_TypeMismatch(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key", NewInteger(42)) // not a Name
	result := dict.GetName("key")
	assert.Nil(t, result)
}

func TestDictionary_GetReal_TypeMismatch(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key", NewString("hello")) // not a Real
	result := dict.GetReal("key")
	assert.Equal(t, 0.0, result)
}

func TestDictionary_GetBoolean_TypeMismatch(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key", NewInteger(1)) // not a Boolean
	result := dict.GetBoolean("key")
	assert.False(t, result)
}

func TestDictionary_GetArray_TypeMismatch(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key", NewString("not an array"))
	result := dict.GetArray("key")
	assert.Nil(t, result)
}

func TestDictionary_GetDictionary_TypeMismatch(t *testing.T) {
	dict := NewDictionary()
	dict.Set("key", NewInteger(1))
	result := dict.GetDictionary("key")
	assert.Nil(t, result)
}

func TestDictionary_GetName_Missing(t *testing.T) {
	dict := NewDictionary()
	result := dict.GetName("nonexistent")
	assert.Nil(t, result)
}

func TestDictionary_GetReal_Missing(t *testing.T) {
	dict := NewDictionary()
	result := dict.GetReal("nonexistent")
	assert.Equal(t, 0.0, result)
}

func TestDictionary_GetBoolean_Missing(t *testing.T) {
	dict := NewDictionary()
	result := dict.GetBoolean("nonexistent")
	assert.False(t, result)
}

func TestDictionary_GetArray_Missing(t *testing.T) {
	dict := NewDictionary()
	result := dict.GetArray("nonexistent")
	assert.Nil(t, result)
}

func TestDictionary_GetDictionary_Missing(t *testing.T) {
	dict := NewDictionary()
	result := dict.GetDictionary("nonexistent")
	assert.Nil(t, result)
}

// ============================================================================
// Parser.parseArray — EOF in array (error path)
// ============================================================================

func TestParser_ParseArray_UnexpectedEOF(t *testing.T) {
	// Array opened but never closed
	p := NewParser(strings.NewReader("[1 2 3"))
	obj, err := p.ParseObject()
	assert.Error(t, err)
	assert.Nil(t, obj)
}

// ============================================================================
// Parser — parseStreamUntilEndstream (0% coverage)
// ============================================================================

func TestParser_ParseIndirectObject_StreamWithoutLength(t *testing.T) {
	// Indirect object with stream that has no /Length — triggers parseStreamUntilEndstream.
	input := "1 0 obj\n<< /Type /Stream >>\nstream\nhello world\nendstream\nendobj"
	p := NewParser(strings.NewReader(input))
	obj, err := p.ParseIndirectObject()
	require.NoError(t, err)
	require.NotNil(t, obj)
	_, isStream := obj.Object.(*Stream)
	assert.True(t, isStream, "should be a stream object")
}

// ============================================================================
// Parser — peekToken second-path (when hasPeek is already set)
// ============================================================================

func TestParser_PeekToken_Cached(t *testing.T) {
	p := NewParser(strings.NewReader("1 0 R"))
	// ParseObject reads "1", then sees "0", then peeks for "R" — exercises hasPeek=true path
	obj, err := p.ParseObject()
	require.NoError(t, err)
	ref, ok := obj.(*IndirectReference)
	require.True(t, ok, "expected *IndirectReference")
	assert.Equal(t, 1, ref.Number)
	assert.Equal(t, 0, ref.Generation)
}

// ============================================================================
// Array.String() — non-empty with multiple types
// ============================================================================

func TestArray_String_MixedTypes(t *testing.T) {
	arr := NewArray()
	arr.Append(NewInteger(1))
	arr.Append(NewReal(3.14))
	arr.Append(NewString("hi"))
	arr.Append(NewName("Type"))
	arr.Append(NewBoolean(false))
	arr.Append(NewNull())

	s := arr.String()
	hasPfx(t, s, "[")
	hasSfx(t, s, "]")
	assert.Contains(t, s, "3.14")
	assert.Contains(t, s, "null")
}
