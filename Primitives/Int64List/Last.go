package Int64List


func(q *query) Last() int64{
	return  q.source[len(q.source)-1]
}