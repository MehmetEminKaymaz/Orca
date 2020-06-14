package StringList


func(q *query) Reverse() {
	for i, j := 0, len(q.source)-1; i < j; i, j = i+1, j-1 {
		q.source[i], q.source[j] = q.source[j], q.source[i]
	}
}
