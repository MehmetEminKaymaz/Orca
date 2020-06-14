package Float64List


func(q *query) CopyTo() []float64{
	var newObj []float64
	newObj=append(newObj,q.source...)
	return newObj
}
