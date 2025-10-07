package dag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"brains/internal/dag"
)

func TestDAGVertexAcyclic(t *testing.T) {
	d, err := dag.NewDAG[int]("root")
	assert.Nil(t, err)

	v1 := &dag.Vertex[int]{Name: "a"}
	v2 := &dag.Vertex[int]{Name: "b"}

	err1 := d.AddVertex(v1)
	err2 := d.AddVertex(v2)
	err3 := d.AddVertex(v1)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.ErrorContains(t, err3, "vertex already exists")
}

func TestDAGVertexAndEdge(t *testing.T) {
	d, err := dag.NewDAG[int]("root")
	assert.Nil(t, err)

	v1 := &dag.Vertex[int]{Name: "a"}
	v2 := &dag.Vertex[int]{Name: "b"}

	_ = d.AddVertex(v1)
	_ = d.AddVertex(v2)

	d.Connect(v1.Name, v2.Name)

	assert.True(t, d.GetVertices()["a"].Name == "a")
	assert.True(t, d.GetVertices()["b"].Name == "b")

	edges, err1 := d.GetEdges()
	assert.Nil(t, err1)
	assert.True(t, edges[0].Target == "b")
}

func TestDAGTopologicalOrder(t *testing.T) {
	d, err := dag.NewDAG[string]("root")
	assert.Nil(t, err)

	a := &dag.Vertex[string]{Name: "a", Run: func(inputs map[string]string) (string, error) { return "a", nil }}
	b := &dag.Vertex[string]{Name: "b", Run: func(inputs map[string]string) (string, error) { return "b", nil }}
	c := &dag.Vertex[string]{Name: "c", Run: func(inputs map[string]string) (string, error) { return "c", nil }}
	d1 := &dag.Vertex[string]{Name: "d", Run: func(inputs map[string]string) (string, error) { return "d", nil }}

	_ = d.AddVertex(a)
	_ = d.AddVertex(b)
	_ = d.AddVertex(c)
	_ = d.AddVertex(d1)

	d.Connect("root", "a")
	d.Connect("a", "b")
	d.Connect("a", "c")
	d.Connect("b", "d")
	d.Connect("c", "d")

	results, err1 := d.Run()
	assert.NoError(t, err1)
	assert.Equal(t, "a", results["a"])
	assert.Equal(t, "b", results["b"])
	assert.Equal(t, "c", results["c"])
	assert.Equal(t, "d", results["d"])
}

func TestDAGResultPropagation(t *testing.T) {
	d, err := dag.NewDAG[int]("root")
	assert.Nil(t, err)

	a := &dag.Vertex[int]{
		Name: "a",
		Run:  func(inputs map[string]int) (int, error) { return 1, nil },
	}
	b := &dag.Vertex[int]{
		Name: "b",
		Run: func(inputs map[string]int) (int, error) {
			return inputs["a"] + 1, nil
		},
	}
	c := &dag.Vertex[int]{
		Name: "c",
		Run: func(inputs map[string]int) (int, error) {
			return inputs["a"] + 2, nil
		},
	}
	d1 := &dag.Vertex[int]{
		Name: "d",
		Run: func(inputs map[string]int) (int, error) {
			return inputs["a"] + inputs["b"] + inputs["c"], nil
		},
	}
	e := &dag.Vertex[int]{
		Name: "e",
		Run: func(inputs map[string]int) (int, error) {
			return inputs["b"] + inputs["c"], nil
		},
	}

	_ = d.AddVertex(a)
	_ = d.AddVertex(b)
	_ = d.AddVertex(c)
	_ = d.AddVertex(d1)
	_ = d.AddVertex(e)

	d.Connect("root", "a")
	d.Connect("a", "b")
	d.Connect("a", "c")
	d.Connect("a", "d")
	d.Connect("b", "d")
	d.Connect("c", "d")
	d.Connect("b", "e")
	d.Connect("c", "e")

	results, err1 := d.Run()
	assert.NoError(t, err1)
	assert.Equal(t, 1, results["a"])
	assert.Equal(t, 2, results["b"])
	assert.Equal(t, 3, results["c"])
	assert.Equal(t, 6, results["d"])
	assert.Equal(t, 5, results["e"])
}

func TestDAGVisualizeComplex(t *testing.T) {
	d, err := dag.NewDAG[int]("root")
	assert.Nil(t, err)

	v1 := &dag.Vertex[int]{Name: "a"}
	v2 := &dag.Vertex[int]{Name: "b"}
	v3 := &dag.Vertex[int]{Name: "c"}

	_ = d.AddVertex(v1)
	_ = d.AddVertex(v2)
	_ = d.AddVertex(v3)

	d.Connect("root", v1.Name)
	d.Connect(v1.Name, v2.Name)
	d.Connect(v2.Name, v3.Name)
	d.Connect(v1.Name, v3.Name)
	d.Visualize()
}

func TestDAGVisualizeSimple(t *testing.T) {
	d, err := dag.NewDAG[int]("root")
	assert.Nil(t, err)

	v1 := &dag.Vertex[int]{Name: "a"}
	v2 := &dag.Vertex[int]{Name: "b"}

	d.Connect("root", v1.Name)
	_ = d.AddVertex(v1)
	_ = d.AddVertex(v2)

	d.Connect(v1.Name, v2.Name)
	d.Visualize()
}
