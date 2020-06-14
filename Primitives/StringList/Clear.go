package StringList


func(q *query) Clear(){
	q.source=q.source[:0]
}

