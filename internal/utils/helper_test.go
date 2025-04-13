package utils_test

import (
	"testing"

	"github.com/aferryc/yars/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestChunkSlice(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		result := utils.ChunkSlice([]int{}, 3)
		assert.Equal(t, [][]int{}, result)
	})

	t.Run("Slice smaller than chunk size", func(t *testing.T) {
		result := utils.ChunkSlice([]int{1, 2}, 3)
		assert.Equal(t, [][]int{{1, 2}}, result)
	})

	t.Run("Slice equal to chunk size", func(t *testing.T) {
		result := utils.ChunkSlice([]int{1, 2, 3}, 3)
		assert.Equal(t, [][]int{{1, 2, 3}}, result)
	})

	t.Run("Slice larger than chunk size", func(t *testing.T) {
		result := utils.ChunkSlice([]int{1, 2, 3, 4, 5, 6, 7}, 3)
		assert.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}, result)
	})

	t.Run("Zero chunk size", func(t *testing.T) {
		result := utils.ChunkSlice([]int{1, 2, 3}, 0)
		assert.Nil(t, result)
	})

	t.Run("Negative chunk size", func(t *testing.T) {
		result := utils.ChunkSlice([]int{1, 2, 3}, -1)
		assert.Nil(t, result)
	})
}
