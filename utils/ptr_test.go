package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDerefPtr(t *testing.T) {

	testCases := []struct {
		ptr      *int
		_default int
		expected int
	}{
		{ptr: (*int)(nil), _default: 1, expected: 1},
		{ptr: NewPtr(2), _default: 1, expected: 2},
	}
	for _, testCase := range testCases {
		require.Equal(t, testCase.expected, DerefPtr(testCase.ptr, testCase._default))
	}
}
