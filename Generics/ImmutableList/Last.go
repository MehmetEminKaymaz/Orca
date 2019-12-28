package ImmutableList


func(q query) Last() (item interface{}){
	return q.v.Index(q.v.Len()-1)
}
