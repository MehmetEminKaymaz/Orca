package LinkedList

func(l *linkedList) Add(item interface{}) *linkedList{
	l.v.PushBack(item)
	return  l
}

