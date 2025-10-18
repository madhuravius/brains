package dag

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/dominikbraun/graph"
	"github.com/muesli/termenv"
)

type SupportedDAGDataTypes interface{ int | string }

func NewDAG[T SupportedDAGDataTypes, D any](rootVertex string) (DAGImpl[T, D], error) {
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
	v.DAG = d
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
		if v.Run == nil {
			continue
		}
		if v.shouldSkip() {
			continue
		}

		inputs := d.collectInputs(name, results)

		result, err := runWithRetry(v, inputs)
		if err != nil {
			return nil, fmt.Errorf("vertex %s failed: %w", name, err)
		}

		results[name] = result
	}

	return results, nil
}

// collectInputs builds the input map for a vertex from prior results.
func (d *DAG[T, D]) collectInputs(target string, results map[string]T) map[string]T {
	inputs := make(map[string]T)
	edges, _ := d.graph.Edges()
	for _, e := range edges {
		if e.Target == target {
			if val, ok := results[e.Source]; ok {
				inputs[e.Source] = val
			}
		}
	}
	return inputs
}

// runWithRetry executes a vertex’s Run function with retry logic.
func runWithRetry[T any, D any](v *Vertex[T, D], inputs map[string]T) (T, error) {
	retries := 1
	if v.EnableRetry {
		retries = v.MaxRetries
		if retries <= 0 {
			retries = DefaultMaxRetries
		}
	}

	var zero T
	for attempt := 1; attempt <= retries; attempt++ {
		result, err := v.Run(inputs)
		if err == nil {
			return result, nil
		}

		if attempt < retries {
			fmt.Printf("Retrying %s (%d/%d): %v\n", v.Name, attempt, retries, err)
		} else {
			return zero, err
		}
	}
	return zero, nil
}

func (d *DAG[T, D]) Visualize() {
	edges, _ := d.graph.Edges()

	adj := make(map[string][]*Vertex[T, D])
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], d.vertices[e.Target])
	}

	// Get topological order
	order, err := graph.TopologicalSort(d.graph)
	if err != nil {
		fmt.Println("error sorting DAG:", err)
		return
	}
	// Assign order to vertices
	counter := 2
	for _, name := range order {
		if name == d.rootVertex.Name {
			continue
		}
		if v, ok := d.vertices[name]; ok {
			v.Order = counter
		}
		counter += 1
	}

	var sb strings.Builder
	sb.WriteString("# DAG Visualization: " + d.rootVertex.Name + "\n\n")
	visited := make(map[string]bool)
	for name := range d.vertices {
		d.visualizeNode(name, adj, visited, &sb)
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

func (d *DAG[T, D]) visualizeNode(name string, adj map[string][]*Vertex[T, D], visited map[string]bool, sb *strings.Builder) {
	if visited[name] {
		return
	}
	visited[name] = true

	if name == d.rootVertex.Name {
		sb.WriteString("1. `" + name + "`")
	} else {
		d.vertices[name].visualizeNonRootVertex(sb)
	}
	children := adj[name]
	if len(children) > 0 {
		sb.WriteString(" **->** ")
		for i, c := range children {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(c.Name)
		}
	}
	sb.WriteString("\n\n")

	for _, c := range children {
		d.visualizeNode(c.Name, adj, visited, sb)
	}
}

func (v *Vertex[T, D]) shouldSkip() bool {
	if v.SkipConfig != nil && v.SkipConfig.Enabled {
		return true
	}

	edges, _ := v.DAG.GetEdges()
	verts := v.DAG.GetVertices()

	parents := make(map[string][]string, len(edges))
	for _, e := range edges {
		parents[e.Target] = append(parents[e.Target], e.Source)
	}

	seen := make(map[string]bool)
	var dfs func(string) bool
	dfs = func(id string) bool {
		if seen[id] {
			return false
		}
		seen[id] = true

		if a := verts[id]; a != nil && a.SkipConfig != nil && a.SkipConfig.Enabled {
			return true
		}
		return slices.ContainsFunc(parents[id], dfs)
	}

	return slices.ContainsFunc(parents[v.Name], dfs)
}

func (v *Vertex[T, D]) visualizeNonRootVertex(sb *strings.Builder) {
	vertexAsString := fmt.Sprintf("%d. ", v.Order)
	if v.shouldSkip() {
		vertexAsString += "`" + v.Name + "`" + "[__SKIPPED__](#)"
	} else {
		vertexAsString += "`" + v.Name + "`"
	}
	sb.WriteString(vertexAsString)
}
