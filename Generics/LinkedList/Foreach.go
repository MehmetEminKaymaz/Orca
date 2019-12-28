package LinkedList

func(l *linkedList) Foreach(do func(x interface{})(y interface{})) *linkedList{


	for e:=l.v.Front();e!=nil;e=e.Next(){
		e.Value=do(e.Value)
	}

	return l



}
