package Float32List

func(q *query) Distinct() {

	var slice []float32
	IsHere:=func(item float32)(state bool){
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