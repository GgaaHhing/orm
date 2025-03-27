package reflect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterateFields(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	testCase := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]any
	}{
		{
			name: "first",
			entity: User{
				Name: "John",
				age:  20,
			},
			wantErr: nil,
			wantRes: map[string]any{
				"Name": "John",
				"age":  0,
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			fields, err := IterateFields(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, fields)
		})
	}
}
