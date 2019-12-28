package LinkedList

import "container/list"

func(l *linkedList) Where(ok func(x interface{}) bool) *linkedList {

	newList:=list.New()
	wrapper := linkedList{
		v:newList,
	}

	for e:=l.v.Front();e!=nil;e=e.Next(){

		if ok(e.Value){
			wrapper.v.PushBack(e.Value)
		}

	}

	l=&wrapper

	return l


}
