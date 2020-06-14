package Float32List

func(q *query) Foreach(do func(x float32) (y float32)){
	for i:=0;i< len(q.source);i++{
		q.source[i]=do(q.source[i])
	}
}