package LinkedList


func(l *linkedList) Any(look func(item interface{}) bool) (state bool){
	state=false

	for e:=l.v.Front();e!=nil;e=e.Next(){
		if look(e.Value){
			state=true
			break
		}
	}

	return

}
