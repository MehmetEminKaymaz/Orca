package LinkedList

import "container/list"

func(l *linkedList) Reverse() *linkedList{

	newList :=list.New()



	for e:=l.v.Front();e!=nil;e=e.Next(){
		newList.PushBack(e.Value)
	}

	wrapper:=linkedList{
		v:newList,
	}

	l=&wrapper

	return l


}