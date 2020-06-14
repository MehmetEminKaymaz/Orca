package StringList


func(q *query) CopyTo() []string{
	var newObj []string
	newObj=append(newObj,q.source...)
	return newObj
}
