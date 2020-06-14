package Float32List

func(q *query) CopyTo() []float32{
    var newObj []float32
    newObj=append(newObj,q.source...)
    return newObj
}
