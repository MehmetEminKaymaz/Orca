package Int32List


func (q *query) AddRange(elems []int32) {
	q.source=append(q.source,elems...)
}
