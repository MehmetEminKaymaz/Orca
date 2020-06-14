package Float64List


func(q *query) Clear(){
	q.source=q.source[:0]
}
