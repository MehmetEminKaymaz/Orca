package StringList


func(q *query) Foreach(do func(x string) (y string)){
	for i:=0;i< len(q.source);i++{
		q.source[i]=do(q.source[i])
	}
}