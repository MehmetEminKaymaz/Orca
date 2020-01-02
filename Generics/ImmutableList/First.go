package ImmutableList


func(q query) First() (item interface{}){

	return q.v.Index(0)

}
