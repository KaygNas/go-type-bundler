package main

type Animal struct {
	LitType       string
	ArrayType     [5]int
	PointerType   *int
	MapType       map[string]int
	SliceType     []int
	ChannelType   chan int
	InterfaceType interface{}
	FunctionType  func()
	StructType    struct {
		LitType       string
		ArrayType     [5]int
		PointerType   *int
		MapType       map[string]int
		SliceType     []int
		ChannelType   chan int
		InterfaceType interface{}
		FunctionType  func()
	}
}
