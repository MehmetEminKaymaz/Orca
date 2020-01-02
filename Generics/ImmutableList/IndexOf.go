package ImmutableList

func(q query) IndexOf(item interface{})(index int){
	index=0
	for i:=0;i<q.v.Len();i++{
		if q.v.Index(i).Interface()==item{
			index=i
			return
		}
	}
	return

}
