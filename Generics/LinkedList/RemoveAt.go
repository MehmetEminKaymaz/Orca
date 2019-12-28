package LinkedList

func(l *linkedList) RemoveAt(index int) *linkedList{

	counter:=0

	for e:=l.v.Front();e!=nil;e=e.Next(){
		if counter==index{
			l.v.Remove(e)
			return l
		}
		counter++
	}
return l
}
