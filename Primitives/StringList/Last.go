package StringList


func(q *query) Last() string{
	return  q.source[len(q.source)-1]
}