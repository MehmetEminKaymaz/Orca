package Int64List


func(q *query) CopyTo() []int64{
	var newObj []int64
	newObj=append(newObj,q.source...)
	return newObj
}