package main

import (
	"brains/internal/dag"

	"github.com/pterm/pterm"
)

func main() {
	d, err := dag.NewDAG[int, int]("_dag")
	if err != nil {
		pterm.Fatal.Printfln("dag.NewDAG: %v", err)
	}

	v1 := &dag.Vertex[int, int]{Name: "a"}
	v2 := &dag.Vertex[int, int]{Name: "b"}
	v3 := &dag.Vertex[int, int]{Name: "c"}

	_ = d.AddVertex(v1)
	_ = d.AddVertex(v2)
	_ = d.AddVertex(v3)

	d.Connect("root", v1.Name)
	d.Connect(v1.Name, v2.Name)
	d.Connect(v2.Name, v3.Name)
	d.Connect(v1.Name, v3.Name)
	d.Visualize()
}
