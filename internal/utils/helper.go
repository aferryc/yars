package utils

func ChunkSlice[T any](items []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return nil
	}

	// If the slice is empty, return empty result
	if len(items) == 0 {
		return [][]T{}
	}

	// Calculate number of chunks
	numChunks := (len(items) + chunkSize - 1) / chunkSize

	// Create the outer slice
	chunks := make([][]T, numChunks)

	// Fill each chunk
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		// Handle the last chunk which might be smaller
		if end > len(items) {
			end = len(items)
		}

		// Create the chunk and copy elements
		chunks[i] = items[start:end]
	}

	return chunks
}
