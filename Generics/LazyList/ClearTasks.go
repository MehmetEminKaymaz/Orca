package LazyList


func (c collection) ClearTasks() collection {

	c.Tasks = c.Tasks[:0]
	c.Items = c.Items[:0]

	return c
}