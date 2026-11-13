package main

import (
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type labelExtractionTestCase struct {
	assemblyStr            string
	expectedLabelsAccessed []string
	expectedLabelDeclared  string
}

func TestLabelExtractor_ExtractLabelsPullsOutTheCorrectLabels(t *testing.T) {
	for _, tc := range []labelExtractionTestCase{
		{".foo", []string{}, "foo"},
		{"add i.foo, $.foo, $0", []string{"foo"}, ""},
		{"mlt i.foo, $.bar, $.biz", []string{"foo", "bar", "biz"}, ""},
		{".data 12", []string{}, "data"},
		{".data i.foo", []string{"foo"}, "data"},
		{".data $.foo", []string{"foo"}, "data"},
		{".data i.foo 12 $.bar", []string{"foo", "bar"}, "data"},
	} {
		line := tsvasmLine{tc.assemblyStr, 12}
		extractor := newLabelExtractor()
		require.NoError(t, extractor.ExtractLabels(line))

		assert.Equal(t, tc.expectedLabelDeclared != "", extractor.LabelDeclared(tc.expectedLabelDeclared), "a label was declared but the extractor found none")

		for _, accessedLabel := range tc.expectedLabelsAccessed {
			assert.True(t, extractor.LabelAccessed(accessedLabel), fmt.Sprintf("'%v' is expected to be accessed", accessedLabel))
		}
	}
}

func TestLabelExtractor_DoesNotAllowMultipleDeclarationsOfALabel(t *testing.T) {
	extractor := newLabelExtractor()
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{".foo", 12}))
	require.Error(t, extractor.ExtractLabels(tsvasmLine{".foo", 24}))
}

func TestLabelExtractor_AcceptsValidLabelNames(t *testing.T) {
	for _, name := range []string{
		".label",
		".data",
		".1234",
		"._1234",
		".my-cool-label",
		".my_cool_label",
		".__foo-bar123__",
	} {
		t.Run(name, func(t *testing.T) {
			extractor := newLabelExtractor()
			require.NoError(t, extractor.ExtractLabels(tsvasmLine{name, 1}), "label name should be valid")

			assert.True(t, extractor.LabelDeclared(name[1:]), "the label name should have been declared")
		})
	}
}

func TestLabelExtractor_RejectsInvalidLabelReferences(t *testing.T) {
	for _, tc := range []tsvasmLine{
		{"add i.$foo, 0, $0", 3},
		{"add i., 0, $0", 3},
		{"add i.foo.bar, 0, $0", 3},
		{"add $. 0, $0, $0", 3},
	} {
		t.Run(tc.assemblyCode, func(t *testing.T) {
			extractor := newLabelExtractor()
			err := extractor.ExtractLabels(tc)

			require.Error(t, err, "the label reference should be invalid")
			assert.Contains(t, err.Error(), "invalid label reference")
		})
	}
}

func TestLabelExtractor_ValidateAccessionsAllowsForwardReferences(t *testing.T) {
	extractor := newLabelExtractor()
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{"jit 1, i.end", 1}))
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{".end", 2}))
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{"hlt", 3}))

	assert.NoError(t, extractor.ValidateAccessions(), "a label referenced before it is declared is fine")
}

func TestLabelExtractor_ValidateAccessionsRejectsUndeclaredLabels(t *testing.T) {
	extractor := newLabelExtractor()
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{"add i.nope, 0, $0", 3}))
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{"hlt", 4}))

	err := extractor.ValidateAccessions()
	require.Error(t, err, "referencing a label that was never declared should fail")
	assert.Contains(t, err.Error(), "nope")
	assert.Contains(t, err.Error(), "3")
}

func TestLabelExtractor_RejectsInvalidLabelDeclarations(t *testing.T) {
	for _, tc := range []struct {
		assemblyCode string
		expectedPart string
	}{
		// A valid label followed by a value that is not a value: a data label with a
		// bad value is still a data label, so the error is about the value...
		{".bad label", "invalid label 'bad'"},
		{".foo bar baz", "invalid label 'foo'"},
		{".foo bar", "invalid label 'foo'"},
		// ...but a valid label followed by non-value junk (no token boundary) is still
		// a label that was not left alone on its line
		{".foo.", "must be alone on its line"},
		// ...as opposed to a name that is invalid to begin with
		{".", "invalid label"},
		{".$foo", "invalid label"},
	} {
		t.Run(tc.assemblyCode, func(t *testing.T) {
			extractor := newLabelExtractor()
			err := extractor.ExtractLabels(tsvasmLine{tc.assemblyCode, 5})

			require.Error(t, err, "the label declaration should be invalid")
			assert.Contains(t, err.Error(), tc.expectedPart, "the error should explain what is wrong")
		})
	}
}

func TestLabelExtractor_RejectsInvalidDataValues(t *testing.T) {
	for _, tc := range []struct {
		assemblyCode string
		expectedPart string
	}{
		{".data foo", "invalid label 'data'"}, // a bare word is not a value
		{".data 1.5", "invalid label 'data'"}, // a float is not a value
		{".data 12,", "invalid label 'data'"}, // no comma tolerance
		{".data $5", "invalid label 'data'"},
		{".data -i12", "invalid label 'data'"},
	} {
		t.Run(tc.assemblyCode, func(t *testing.T) {
			extractor := newLabelExtractor()
			err := extractor.ExtractLabels(tsvasmLine{tc.assemblyCode, 5})

			require.Error(t, err, "the data value should be invalid")
			assert.Contains(t, err.Error(), tc.expectedPart, "the error should name the label")
			assert.Contains(t, err.Error(), "not a valid data value", "the error should explain what a data value is")
		})
	}
}

func TestLabelExtractor_RejectsDuplicateDataLabelDeclarations(t *testing.T) {
	extractor := newLabelExtractor()
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{".foo", 12}))
	require.Error(t, extractor.ExtractLabels(tsvasmLine{".foo 12", 24}), "a data label may not redeclare a name")

	extractor = newLabelExtractor()
	require.NoError(t, extractor.ExtractLabels(tsvasmLine{".foo 12", 12}))
	require.Error(t, extractor.ExtractLabels(tsvasmLine{".foo", 24}), "a regular label may not redeclare a data label")
}

func TestLabelExtractor_ValidateAccessionsChecksDataValueReferences(t *testing.T) {
	t.Run("a reference to a declared label is fine, even forward", func(t *testing.T) {
		extractor := newLabelExtractor()
		require.NoError(t, extractor.ExtractLabels(tsvasmLine{".data $.foo", 1}))
		require.NoError(t, extractor.ExtractLabels(tsvasmLine{".foo", 2}))

		assert.NoError(t, extractor.ValidateAccessions(), "a label referenced in a data value and declared later is fine")
	})

	t.Run("a reference to a never-declared label fails", func(t *testing.T) {
		extractor := newLabelExtractor()
		require.NoError(t, extractor.ExtractLabels(tsvasmLine{".data $.nope", 3}))

		err := extractor.ValidateAccessions()

		require.Error(t, err, "referencing a label that was never declared should fail")
		assert.Contains(t, err.Error(), "nope")
		assert.Contains(t, err.Error(), "3")
	})
}
