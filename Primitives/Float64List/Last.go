package Float64List


func(q *query) Last() float64{
	return  q.source[len(q.source)-1]
}

