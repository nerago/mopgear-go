package neragogearing

type TestEnum int8

const (
	AAA TestEnum = iota
	BBB TestEnum = iota
	CCC TestEnum = iota
)

type TestItem struct {
	id    uint32
	value uint32
}

type TestItem2 struct {
	TestItem
}

type TestMe struct {
	blah float64
	list []TestItem
}
