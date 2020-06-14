package Int64List


func(q *query) ElementAt(index int) int64{
	return q.source[index]
}
