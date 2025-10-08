package dag

import "github.com/dominikbraun/graph"

type Vertex[T any, D any] struct {
	Name string
	Run  func(inputs map[string]T) (T, error)
	DAG  *DAG[T, D]
}

type DAG[T any, D any] struct {
	graph      graph.Graph[string, *Vertex[T, D]]
	rootVertex *Vertex[T, D]
	vertices   map[string]*Vertex[T, D]

	data D
}
