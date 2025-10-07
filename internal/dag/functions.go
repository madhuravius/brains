package dag

import (
	"fmt"

	"github.com/dominikbraun/graph"
)

type SupportedDAGDataTypes interface{ int | string }

func NewDAG[T SupportedDAGDataTypes](rootVertex string) *DAG[T] {
	vertexHash := func(v *Vertex[T]) string {
		return v.Name
	}

	g := graph.New(vertexHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles())
	return &DAG[T]{graph: g}
}

func (d *DAG[T]) AddVertex(v *Vertex[T]) error {
	return d.graph.AddVertex(v)
}

func (d *DAG[T]) Connect(src, dest string) {
	_ = d.graph.AddEdge(src, dest)
}

func (d *DAG[T]) Run() (map[string]T, error) {
	results := make(map[string]T)

	order, err := graph.TopologicalSort(d.graph)
	if err != nil {
		return nil, fmt.Errorf("cannot sort DAG: %w", err)
	}

	for _, name := range order {
		v, _ := d.graph.Vertex(name)

		inputs := map[string]T{}
		edges, _ := d.graph.Edges()
		for _, e := range edges {
			if e.Target == name {
				if val, ok := results[e.Source]; ok {
					inputs[e.Source] = val
				}
			}
		}

		result, err := v.Run(inputs)
		if err != nil {
			return nil, fmt.Errorf("error running vertex %s: %w", name, err)
		}
		results[name] = result
	}

	return results, nil
}
