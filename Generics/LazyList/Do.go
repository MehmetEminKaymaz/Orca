package LazyList



func (c collection) Do() collection { //call all operators

	copyC := c.Tasks
	for itemOfFunk, task := range copyC {

		switch task.KindOfTask { //one case for each different function

		case "Void":
			funk := task.Do().(func(*collection, interface{}))
			funk(&c, c.Items[itemOfFunk])

		case "Void2":
			funk := task.Do().(func(*collection))
			funk(&c)
		case "Func":
			funk := task.Do().(func(*collection, func(interface{}) interface{}))
			funk(&c, c.Items[itemOfFunk].(func(interface{}) interface{}))
		case "Bool":
			funk := task.Do().(func(*collection, func(interface{}) bool))
			funk(&c, c.Items[itemOfFunk].(func(interface{}) bool))

		default:

		}

	}
	c.Tasks = c.Tasks[:0] //delete executed task from tasks
	c.Items = c.Items[:0] //delete items after execution

	return c

}
