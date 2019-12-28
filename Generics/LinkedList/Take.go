package LinkedList

import "container/list"

func(l *linkedList) Take(tk int) *linkedList {

	newList:=list.New()
	wrapper:=linkedList{
		v:newList,
	}
	counter:=0
	for e:=l.v.Front();e!=nil;e=e.Next(){

		if counter<=tk{
			wrapper.v.PushBack(e.Value)
		}

		counter++
	}

	l=&wrapper


	return l


}