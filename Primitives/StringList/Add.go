package StringList


func (q *query) Add(elem string) {
	q.source=append(q.source,elem)
}