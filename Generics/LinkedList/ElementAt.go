package LinkedList

func(l *linkedList) ElementAt(index int) (x interface{}){
	counter:=0
	for e:=l.v.Front();e!=nil;e=e.Next(){

		if counter==index{
			return e.Value
		}

		counter++
	}

	return nil



}
