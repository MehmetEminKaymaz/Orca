package Float32List

func(q *query) Any(look func(i float32) bool) (state bool){
	state=false
	for i:=0 ;i< len(q.source);i++{
		if look(q.source[i]){
			state=true
			return
		}
	}
	return
}
