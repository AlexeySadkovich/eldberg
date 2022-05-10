package discover

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasBit(t *testing.T) {
	testCases := []struct {
		b   byte
		pos uint
		out bool
	}{
		{
			b:   1, // 00000001
			pos: 7, // 01234567 - positions
			out: true,
		},
		{
			b:   2, // 00000010
			pos: 6,
			out: true,
		},
		{
			b:   3, // 00000011
			pos: 6,
			out: true,
		},
		{
			b:   3, // 00000011
			pos: 7,
			out: true,
		},
		{
			b:   4, // 00000100
			pos: 5,
			out: true,
		},
	}

	for _, tc := range testCases {
		res := HasBit(tc.b, tc.pos)
		require.Equal(t, tc.out, res)
	}
}
