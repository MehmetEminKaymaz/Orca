package LinkedList

import "reflect"

func(l *linkedList) Contains(item interface{}) (state bool){


	state=false

	for e:=l.v.Front();e!=nil;e=e.Next(){
		if reflect.ValueOf(item).Interface()==e.Value{
			state=true
			return
		}
	}

	return

}
