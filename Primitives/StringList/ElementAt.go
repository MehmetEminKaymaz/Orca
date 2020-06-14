package StringList


func(q *query) ElementAt(index int) string{
	return q.source[index]
}

