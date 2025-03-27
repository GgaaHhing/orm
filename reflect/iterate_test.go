package reflect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterateArray(t *testing.T) {
	testCase := []struct {
		name    string
		entity  any
		wantRes []any
		wantErr error
	}{
		{
			name:    "[]int",
			entity:  [3]int{1, 2, 3},
			wantRes: []any{1, 2, 3},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateArrayOrSlice(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
