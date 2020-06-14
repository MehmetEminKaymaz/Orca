package Float32List

func(q *query) Last() float32{
	return  q.source[len(q.source)-1]
}
