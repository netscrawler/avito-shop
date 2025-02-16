package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMerch(t *testing.T) {
	tests := []struct {
		name      string
		merchName string
		price     uint64
		want      *Merch
	}{
		{
			name:      "создание нового товара",
			merchName: "test-item",
			price:     100,
			want: &Merch{
				Name:  "test-item",
				Price: 100,
			},
		},
		{
			name:      "создание бесплатного товара",
			merchName: "free-item",
			price:     0,
			want: &Merch{
				Name:  "free-item",
				Price: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMerch(tt.merchName, tt.price)
			assert.Equal(t, tt.want, got)
		})
	}
}
