package Int32List


func(q *query) Foreach(do func(x int32) (y int32)){
	for i:=0;i< len(q.source);i++{
		q.source[i]=do(q.source[i])
	}
}