package Float64List


func(q *query) Contains(item float64) (state bool){
	state=false
	for i:=0;i< len(q.source);i++{
		if item==q.source[i]{
			state=true
			return
		}
	}
	return false
}