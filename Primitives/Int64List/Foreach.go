package Int64List


func(q *query) Foreach(do func(x int64) (y int64)){
	for i:=0;i< len(q.source);i++{
		q.source[i]=do(q.source[i])
	}
}
