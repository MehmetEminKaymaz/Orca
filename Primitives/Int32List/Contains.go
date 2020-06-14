package Int32List


func(q *query) Contains(item int32) (state bool){
	state=false
	for i:=0;i< len(q.source);i++{
		if item==q.source[i]{
			state=true
			return
		}
	}
	return false
}