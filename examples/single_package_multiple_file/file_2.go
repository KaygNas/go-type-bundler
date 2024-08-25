package main

type Dog struct {
	StructType struct {
		Name string
	}
	ArrayType     [5]int
	PointerType   *int
	FunctionType  func()
	InterfaceType interface{}
	MapType       map[string]int
	ChannelType   chan int
	SliceType     []int
}
