package main

import (
	"github.com/madhuravius/brains/internal/dag"

	"github.com/pterm/pterm"
)

func regularDAG() {
	d1, err := dag.NewDAG[int, int]("_dag unskipped")
	if err != nil {
		pterm.Fatal.Printfln("dag.NewDAG: %v", err)
	}

	v1 := &dag.Vertex[int, int]{Name: "a"}
	v2 := &dag.Vertex[int, int]{Name: "b"}
	v3 := &dag.Vertex[int, int]{Name: "c"}

	_ = d1.AddVertex(v1)
	_ = d1.AddVertex(v2)
	_ = d1.AddVertex(v3)

	d1.Connect("root (no skip)", v1.Name)
	d1.Connect(v1.Name, v2.Name)
	d1.Connect(v2.Name, v3.Name)
	d1.Connect(v1.Name, v3.Name)
	d1.Visualize()
}

func skippedDAG() {
	d2, err := dag.NewDAG[int, int]("_dag with skips")
	if err != nil {
		pterm.Fatal.Printfln("dag.NewDAG: %v", err)
	}

	v1 := &dag.Vertex[int, int]{Name: "a"}
	v2 := &dag.Vertex[int, int]{Name: "b", SkipConfig: &dag.SkipVertexConfig{Enabled: true, Reason: "skipping for debug"}}
	v3 := &dag.Vertex[int, int]{Name: "c"}
	v4 := &dag.Vertex[int, int]{Name: "d"}

	_ = d2.AddVertex(v1)
	_ = d2.AddVertex(v2)
	_ = d2.AddVertex(v3)
	_ = d2.AddVertex(v4)

	d2.Connect("root (no skip)", v1.Name)
	d2.Connect(v1.Name, v2.Name)
	d2.Connect(v2.Name, v3.Name)
	d2.Connect(v1.Name, v3.Name)
	d2.Connect(v3.Name, v4.Name)
	d2.Visualize()
}

func main() {
	regularDAG()
	skippedDAG()
}
