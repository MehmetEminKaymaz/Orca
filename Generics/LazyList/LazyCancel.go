package LazyList


func (c collection) LazyCancel(count int) collection { //it cancels the last (count) tasks!

	t := task{
		KindOfTask: "Void2",
		Do: func() interface{} {
			return func(col *collection) {
				c.Tasks = c.Tasks[:len(c.Tasks)-count] //clear tasks
				c.Items = c.Items[:len(c.Items)-count] //clear item of tasks
			}
		},
	}

	c.Tasks = append(c.Tasks, t)
	c.Items = append(c.Items, empty{})

	return c
}