package Orca

func(q *Query) First() (item interface{}){

	return q.v.Index(0)

}
