package Int64List


func(q *query) Distinct() {

	var slice []int64
	IsHere:=func(item int64)(state bool){
		state =false
		for k:=0;k< len(slice);k++{
			if slice[k]==item{
				state=true
			}
		}
		return
	}

	for i:=0;i< len(q.source);i++{
		if !IsHere(q.source[i]){
			slice=append(slice,q.source[i])
		}
	}

	q.source=slice

}
