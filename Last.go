package Orca

func(q *Query) Last() (item interface{}){
	return q.v.Index(q.v.Len()-1)
}
