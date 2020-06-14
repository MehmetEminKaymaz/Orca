package StringList


func(q *query) RemoveAt(i int) {

	q.source[len(q.source)-1], q.source[i] = q.source[i], q.source[len(q.source)-1]
	q.source=q.source[:len(q.source)-1]

}
