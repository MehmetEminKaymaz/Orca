package Int32List


func(q *query) CopyTo() []int32{
	var newObj []int32
	newObj=append(newObj,q.source...)
	return newObj
}
