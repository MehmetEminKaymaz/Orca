package Float64List


func (q *query) AddRange(elems []float64) {
	q.source=append(q.source,elems...)
}
