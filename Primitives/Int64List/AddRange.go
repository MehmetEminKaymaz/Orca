package Int64List


func (q *query) AddRange(elems []int64) {
	q.source=append(q.source,elems...)
}
