package StringList



func(q *query) Exist(slice []string)  (state bool){

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