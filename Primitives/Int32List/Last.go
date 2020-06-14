package Int32List


func(q *query) Last() int32{
	return  q.source[len(q.source)-1]
}
