package Float64List



func(q *query) Foreach(do func(x float64) (y float64)){
	for i:=0;i< len(q.source);i++{
		q.source[i]=do(q.source[i])
	}
}
