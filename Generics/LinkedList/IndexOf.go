package LinkedList


func(l *linkedList) IndexOf( item interface{})(index int){
	index=0
	counter:=0
	for e:=l.v.Front();e!=nil;e=e.Next(){
		if e.Value==item{
			index=counter
			return
		}
		counter++
	}

	return
}