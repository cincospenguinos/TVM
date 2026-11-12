package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataValueParserAcceptsTheFourDeclaredForms(t *testing.T) {
	for _, tc := range []struct {
		name          string
		token         string
		expectedKind  dataValueKind
		expectedValue int
		expectedLabel string
	}{
		{name: "a plain number", token: "12", expectedKind: dataValueLiteral, expectedValue: 12},
		{name: "a negative number", token: "-1", expectedKind: dataValueLiteral, expectedValue: -1},
		{name: "the immediate prefix", token: "i12", expectedKind: dataValueLiteral, expectedValue: 12},
		{name: "the immediate prefix with a sign", token: "i-12", expectedKind: dataValueLiteral, expectedValue: -12},
		{name: "an immediate reference", token: "i.bar", expectedKind: dataValueIndexRef, expectedLabel: "bar"},
		{name: "an address reference", token: "$.bar", expectedKind: dataValueDeref, expectedLabel: "bar"},
		{name: "a reference whose name begins with a digit", token: "i.1234", expectedKind: dataValueIndexRef, expectedLabel: "1234"},
		{name: "a reference with dashes and underscores", token: "$.my-label_2", expectedKind: dataValueDeref, expectedLabel: "my-label_2"},
		{name: "a dot after 'i' always begins a reference, even to a signed name", token: "i.-1", expectedKind: dataValueIndexRef, expectedLabel: "-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := parseDataValue(tc.token)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedKind, value.kind)
			assert.Equal(t, tc.expectedValue, value.value)
			assert.Equal(t, tc.expectedLabel, value.label)
		})
	}
}

func TestDataValueParserRejectsAnythingElse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "a bare word", token: "foo"},
		{name: "a float", token: "1.5"},
		{name: "a trailing comma (no tolerance, unlike instruction parameters)", token: "12,"},
		{name: "a sign before the 'i'", token: "-i12"},
		{name: "a bare prefix", token: "i"},
		{name: "an empty reference name", token: "i."},
		{name: "an empty address reference name", token: "$."},
		{name: "an address indicator without a reference", token: "$5"},
		{name: "a register", token: "t0"},
		{name: "a hex number", token: "0x10"},
		{name: "a number with a dot in it", token: "i1.2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseDataValue(tc.token)

			require.Error(t, err, "the token should not be a valid data value")
			assert.Contains(t, err.Error(), "not a valid data value")
			assert.Contains(t, err.Error(), tc.token)
		})
	}
}

func TestParseDataLabel(t *testing.T) {
	for _, tc := range []struct {
		name           string
		line           string
		expectedName   string
		expectedValues []dataValue
	}{
		{
			name:           "a single value",
			line:           ".data 12",
			expectedName:   "data",
			expectedValues: []dataValue{{kind: dataValueLiteral, value: 12}},
		},
		{
			name:           "the name is the entire first chunk, dashes included",
			line:           ".data-12 12",
			expectedName:   "data-12",
			expectedValues: []dataValue{{kind: dataValueLiteral, value: 12}},
		},
		{
			name:           "multiple values",
			line:           ".data 1 2 3",
			expectedName:   "data",
			expectedValues: []dataValue{{kind: dataValueLiteral, value: 1}, {kind: dataValueLiteral, value: 2}, {kind: dataValueLiteral, value: 3}},
		},
		{
			name:         "arbitrary mixes of the forms, in order",
			line:         ".crazy-data__69420 i.bar 12 3 $.biz i.baz",
			expectedName: "crazy-data__69420",
			expectedValues: []dataValue{
				{kind: dataValueIndexRef, label: "bar"},
				{kind: dataValueLiteral, value: 12},
				{kind: dataValueLiteral, value: 3},
				{kind: dataValueDeref, label: "biz"},
				{kind: dataValueIndexRef, label: "baz"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parseDataLabel(tc.line)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedName, parsed.name)
			assert.Equal(t, tc.expectedValues, parsed.values)
		})
	}
}

func TestParseDataLabelRejectsMalformedLines(t *testing.T) {
	for _, tc := range []struct {
		name            string
		line            string
		expectedErrPart string
	}{
		{name: "a bare word is not a value", line: ".data foo", expectedErrPart: "invalid label 'data'"},
		{name: "a float is not a value", line: ".data 1.5", expectedErrPart: "invalid label 'data'"},
		{name: "a trailing comma is not a value", line: ".data 12,", expectedErrPart: "invalid label 'data'"},
		{name: "a name with no values is not a data label", line: ".data", expectedErrPart: "not a data label"},
		{name: "an invalid first chunk is an invalid label", line: ".foo. 12", expectedErrPart: "invalid label"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseDataLabel(tc.line)

			require.Error(t, err, "the line should not be a valid data label")
			assert.Contains(t, err.Error(), tc.expectedErrPart)
		})
	}
}

func TestIsDataLabel(t *testing.T) {
	for _, tc := range []struct {
		line     string
		expected bool
	}{
		{line: ".data 12", expected: true},
		{line: ".data-12 12", expected: true},
		{line: ".data 1 2 3", expected: true},
		{line: ".data i.bar $.baz", expected: true},
		{line: ".foo", expected: false},  // a label alone on its line is a regular declaration
		{line: ".foo.", expected: false}, // no token boundary after the name
		{line: ".foo. bar", expected: false},
		{line: ".", expected: false},
		{line: ".$foo", expected: false},
		{line: "add 1 2 3", expected: false}, // not a label line at all
	} {
		t.Run(tc.line, func(t *testing.T) {
			assert.Equal(t, tc.expected, isDataLabel(tsvasmLine{tc.line, 1}))
		})
	}
}

func TestDataLabelAppendWords(t *testing.T) {
	t.Run("literals and index references resolve immediately", func(t *testing.T) {
		dataLabel := dataLabel{
			name:   "data",
			values: []dataValue{{kind: dataValueLiteral, value: 12}, {kind: dataValueIndexRef, label: "bar"}},
		}

		program, pending, err := dataLabel.appendWords(map[string]int{"bar": 5}, []int{})
		require.NoError(t, err)

		assert.Equal(t, []int{12, 5}, program)
		assert.Empty(t, pending)
	})

	t.Run("a dereference emits a placeholder and a pending reference", func(t *testing.T) {
		dataLabel := dataLabel{
			name:   "data",
			values: []dataValue{{kind: dataValueDeref, label: "foo"}},
		}

		program, pending, err := dataLabel.appendWords(map[string]int{"foo": 9}, []int{})
		require.NoError(t, err)

		assert.Equal(t, []int{0}, program)
		assert.Equal(t, []pendingDataReference{{slot: 0, label: "foo"}}, pending)
	})

	t.Run("the placeholder's slot is its position in the running program", func(t *testing.T) {
		dataLabel := dataLabel{
			name:   "data",
			values: []dataValue{{kind: dataValueLiteral, value: 12}, {kind: dataValueDeref, label: "foo"}},
		}

		program, pending, err := dataLabel.appendWords(map[string]int{"foo": 3}, []int{9, 1101})
		require.NoError(t, err)

		assert.Equal(t, []int{9, 1101, 12, 0}, program)
		assert.Equal(t, []pendingDataReference{{slot: 3, label: "foo"}}, pending)
	})

	t.Run("an unresolvable reference is an error", func(t *testing.T) {
		for _, kind := range []dataValueKind{dataValueIndexRef, dataValueDeref} {
			dataLabel := dataLabel{
				name:   "data",
				values: []dataValue{{kind: kind, label: "nope"}},
			}

			_, _, err := dataLabel.appendWords(map[string]int{}, []int{})

			require.Error(t, err, "a reference to an undeclared label should not resolve")
			assert.Contains(t, err.Error(), "nope")
		}
	})
}

func TestResolveDataReferences(t *testing.T) {
	for _, tc := range []struct {
		name            string
		program         []int
		labelIndexes    map[string]int
		pending         []pendingDataReference
		expected        []int
		expectedErrPart []string
	}{
		{
			name:         "a backward reference copies the already-emitted word",
			program:      []int{0, 9},
			labelIndexes: map[string]int{"foo": 1},
			pending:      []pendingDataReference{{slot: 0, label: "foo"}},
			expected:     []int{9, 9},
		},
		{
			name:         "a forward reference is finalized from its target word",
			program:      []int{0, 1101, 1, 1, 0, 9},
			labelIndexes: map[string]int{"foo": 5},
			pending:      []pendingDataReference{{slot: 0, label: "foo"}},
			expected:     []int{9, 1101, 1, 1, 0, 9},
		},
		{
			name:         "a chain is walked to the first finalized word",
			program:      []int{0, 42, 9},
			labelIndexes: map[string]int{"one": 0, "two": 1},
			pending:      []pendingDataReference{{slot: 0, label: "two"}},
			expected:     []int{42, 42, 9},
		},
		{
			name:         "every link of a chain takes the same value",
			program:      []int{0, 0, 42, 9},
			labelIndexes: map[string]int{"one": 0, "two": 1, "three": 2},
			pending:      []pendingDataReference{{slot: 0, label: "two"}, {slot: 1, label: "three"}},
			expected:     []int{42, 42, 42, 9},
		},
		{
			name:            "a cycle is an error naming the labels involved",
			program:         []int{0, 0},
			labelIndexes:    map[string]int{"a": 0, "b": 1},
			pending:         []pendingDataReference{{slot: 0, label: "b"}, {slot: 1, label: "a"}},
			expectedErrPart: []string{"cycle", "a", "b"},
		},
		{
			name:            "a self-reference is a cycle",
			program:         []int{0},
			labelIndexes:    map[string]int{"self": 0},
			pending:         []pendingDataReference{{slot: 0, label: "self"}},
			expectedErrPart: []string{"cycle", "self"},
		},
		{
			name:            "a label past the end of the program is an error",
			program:         []int{0},
			labelIndexes:    map[string]int{"baz": 1},
			pending:         []pendingDataReference{{slot: 0, label: "baz"}},
			expectedErrPart: []string{"past the end"},
		},
		{
			name:     "no pending references is a no-op",
			program:  []int{9},
			expected: []int{9},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			program := append([]int{}, tc.program...)
			err := resolveDataReferences(program, tc.labelIndexes, tc.pending)

			if len(tc.expectedErrPart) > 0 {
				require.Error(t, err, "the references should not resolve")
				for _, part := range tc.expectedErrPart {
					assert.Contains(t, err.Error(), part)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, program)
		})
	}
}
