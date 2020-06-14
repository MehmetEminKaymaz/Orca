package Float32List

func(q *query) Clear(){
	q.source=q.source[:0]
}
