package domain

type Merch struct {
	Id    int
	Name  string
	Price uint64
}

func NewMerch(name string, price uint64) *Merch {
	return &Merch{
		Name:  name,
		Price: price,
	}
}
