package LazyList

func (c collection) Cancel(count int) collection { //it cancels the last (count) tasks!

	c.Tasks = c.Tasks[:len(c.Tasks)-count] //clear tasks
	c.Items = c.Items[:len(c.Items)-count] //clear item of tasks

	return c
}

func (c collection) GetTaskCount() int { //it returns task count
	return len(c.Tasks)
}