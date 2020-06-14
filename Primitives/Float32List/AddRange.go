package Float32List


func (q *query) AddRange(elems []float32) {
	q.source=append(q.source,elems...)
}
