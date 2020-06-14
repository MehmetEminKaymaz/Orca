package Int64List


func(q *query) Any(look func(i int64) bool) (state bool){
	state=false
	for i:=0 ;i< len(q.source);i++{
		if look(q.source[i]){
			state=true
			return
		}
	}
	return
}