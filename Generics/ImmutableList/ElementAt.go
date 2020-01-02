package ImmutableList

func(q query) ElementAt(index  int ) (x interface{}){
	return q.v.Index(index).Interface()
}
