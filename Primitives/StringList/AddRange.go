package StringList


func (q *query) AddRange(elems []string) {
	q.source=append(q.source,elems...)
}

