package Float32List

func(q *query) ElementAt(index int) float32{
	return q.source[index]
}
