package Orca



func(q *Query) ElementAt(index  int ) (x interface{}){
	return q.v.Index(index).Interface()
}
