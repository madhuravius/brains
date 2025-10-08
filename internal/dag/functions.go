package dag

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/dominikbraun/graph"
	"github.com/muesli/termenv"
)

type SupportedDAGDataTypes interface{ int | string }

func NewDAG[T SupportedDAGDataTypes, D any](rootVertex string) (*DAG[T, D], error) {
	vertexHash := func(v *Vertex[T, D]) string {
		return v.Name
	}

	g := graph.New(vertexHash, graph.Directed(), graph.Acyclic(), graph.PreventCycles())
	dag := DAG[T, D]{graph: g, vertices: make(map[string]*Vertex[T, D])}

	root := &Vertex[T, D]{Name: rootVertex}
	err := dag.AddVertex(root)
	if err != nil {
		return nil, err
	}
	dag.rootVertex = root

	return &dag, nil
}

func (d *DAG[T, D]) AddVertex(v *Vertex[T, D]) error {
	if err := d.graph.AddVertex(v); err != nil {
		return err
	}
	d.vertices[v.Name] = v
	return nil
}

func (d *DAG[T, D]) GetVertices() map[string]*Vertex[T, D] {
	return d.vertices
}

func (d *DAG[T, D]) Connect(src, dest string) {
	_ = d.graph.AddEdge(src, dest)
}

func (d *DAG[T, D]) GetEdges() ([]graph.Edge[string], error) {
	return d.graph.Edges()
}

func (d *DAG[T, D]) Run() (map[string]T, error) {
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

		if v.Run != nil {
			result, err := v.Run(inputs)
			if err != nil {
				return nil, fmt.Errorf("error running vertex %s: %w", name, err)
			}
			results[name] = result
		}
	}

	return results, nil
}

func (d *DAG[T, D]) Visualize() {
	edges, _ := d.graph.Edges()

	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
	}

	var sb strings.Builder
	sb.WriteString("# DAG Visualization: " + d.rootVertex.Name + "\n\n")
	visited := make(map[string]bool)
	for name := range d.vertices {
		visualizeNode(name, adj, visited, &sb)
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
		glamour.WithColorProfile(termenv.TrueColor),
	)
	sortedOut := strings.Split(sb.String(), "\n")
	sort.Strings(sortedOut)
	out, _ := r.Render(strings.Join(sortedOut, "\n\n"))
	fmt.Print(out)
}

func visualizeNode(name string, adj map[string][]string, visited map[string]bool, sb *strings.Builder) {
	if visited[name] {
		return
	}
	visited[name] = true

	sb.WriteString("`" + name + "`")
	children := adj[name]
	if len(children) > 0 {
		sb.WriteString(" **->** ")
		for i, c := range children {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(c)
		}
	}
	sb.WriteString("\n\n")

	for _, c := range children {
		visualizeNode(c, adj, visited, sb)
	}
}
