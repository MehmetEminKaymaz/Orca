package Float32List

func (q *query) Add(elem float32) {
	q.source=append(q.source,elem)
}