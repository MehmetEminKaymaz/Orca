package MutableList

func(q *query) Last() (item interface{}){
	return q.v.Index(q.v.Len()-1)
}
