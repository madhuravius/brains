package dag

import "github.com/dominikbraun/graph"

type SkipVertexConfig struct {
	Enabled bool
	Reason  string
}

type Vertex[T any, D any] struct {
	Name        string
	Order       int
	Run         func(inputs map[string]T) (T, error)
	DAG         DAGImpl[T, D]
	Needs       map[*Vertex[T, D]]bool
	EnableRetry bool
	MaxRetries  int
	SkipConfig  *SkipVertexConfig
}

type DAG[T any, D any] struct {
	graph      graph.Graph[string, *Vertex[T, D]]
	rootVertex *Vertex[T, D]
	vertices   map[string]*Vertex[T, D]
}

type DAGImpl[T any, D any] interface {
	AddVertex(v *Vertex[T, D]) error
	Connect(src, dest string)
	GetEdges() ([]graph.Edge[string], error)
	GetVertices() map[string]*Vertex[T, D]
	Run() (map[string]T, error)
	Visualize()
}
