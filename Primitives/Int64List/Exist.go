package Int64List


func(q *query) Exist(slice []int64)  (state bool){

	state =true
	for i:=0;i< len(slice);i++{
		for k:=0; k< len(q.source);k++{
			if q.source[k]!=slice[i]{
				state=false
				return
			}
		}
	}
	return

}