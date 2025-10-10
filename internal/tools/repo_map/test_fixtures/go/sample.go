package main

const Pi = 3.14

var version = "v1.0"

type Person struct {
	Name string
}

type Greeter interface {
	Greet() string
}

func SayHello(name string) string {
	return "Hello " + name
}

func (p *Person) Greet() string {
	return "Hi, I'm " + p.Name
}
