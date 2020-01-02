package LinkedList


func (l *linkedList) Clear() *linkedList {
	l.v.Init()

	return l
}
