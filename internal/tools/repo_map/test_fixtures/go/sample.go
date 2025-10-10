package sample

// Person represents a person.
type Person struct {
	Name string
}

// Greeter can greet by name.
type Greeter interface {
	Greet(name string) string
}

// Pi is the ratio of a circle's circumference to its diameter.
const Pi = 3.14

// version of the module.
var version = "1.0" //nolint:unused

// SayHello says hi
func SayHello(name string, times int) string {
	return ""
}

// Greet is the Person greeter.
func (p Person) Greet(name string) string {
	return "hi"
}
