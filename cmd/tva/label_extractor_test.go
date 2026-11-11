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

func TestLabelExtractor_RejectsInvalidLabelDeclarations(t *testing.T) {
	for _, tc := range []struct {
		assemblyCode string
		expectedPart string
	}{
		// A valid label followed by extraneous content...
		{".bad label", "must be alone on its line"},
		{".foo bar baz", "must be alone on its line"},
		{".foo.", "must be alone on its line"},
		// ...as opposed to a name that is invalid to begin with
		{".", "invalid label"},
		{".$foo", "invalid label"},
		{".foo bar", "must be alone on its line"},
	} {
		t.Run(tc.assemblyCode, func(t *testing.T) {
			extractor := newLabelExtractor()
			err := extractor.ExtractLabels(tsvasmLine{tc.assemblyCode, 5})

			require.Error(t, err, "the label declaration should be invalid")
			assert.Contains(t, err.Error(), tc.expectedPart, "the error should explain what is wrong")
		})
	}
}
