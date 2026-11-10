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
