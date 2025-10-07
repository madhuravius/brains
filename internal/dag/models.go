package dag

import "github.com/dominikbraun/graph"

type Vertex[T any] struct {
	Name string
	Run  func(inputs map[string]T) (T, error)
}

type DAG[T any] struct {
	graph      graph.Graph[string, *Vertex[T]]
	rootVertex *Vertex[T]
	vertices   map[string]*Vertex[T]
}
