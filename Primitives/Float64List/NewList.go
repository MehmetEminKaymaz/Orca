package Float64List



type query struct {
	source []float64
}

func NewList() query{
	var source []float64
	return query{
		source: source,
	}
}

func NewListFrom(s []float64) query{
	return query{
		source: s,
	}
}
