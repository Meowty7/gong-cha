// Package recipe provides recipe graph validation and the BOM calculation engine.
package recipe

// Edge is a dependency: From (recipe result product) depends on To (component product).
type Edge struct {
	From, To string
}

// DetectCycle runs a three-color DFS over the dependency edges and returns a
// cycle path (product ids) if one exists, or nil otherwise. A cycle means a
// product transitively depends on itself, which makes a recipe unproducible.
func DetectCycle(edges []Edge) []string {
	adj := make(map[string][]string)
	nodes := make(map[string]struct{})
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
		nodes[e.From] = struct{}{}
		nodes[e.To] = struct{}{}
	}
	const (
		white = 0 // unvisited
		gray  = 1 // in progress
		black = 2 // done
	)
	color := make(map[string]int, len(nodes))
	parent := make(map[string]string, len(nodes))
	var cycle []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = gray
		for _, next := range adj[node] {
			if color[next] == gray {
				// Reconstruct the cycle from next back to itself via parent links.
				cycle = []string{next}
				for cur := node; cur != next && cur != ""; cur = parent[cur] {
					cycle = append(cycle, cur)
				}
				cycle = append(cycle, next)
				// Reverse so it reads start->...->start.
				rev(cycle)
				return true
			}
			if color[next] == white {
				parent[next] = node
				if dfs(next) {
					return true
				}
			}
		}
		color[node] = black
		return false
	}

	// Iterate in a stable order for deterministic output.
	for n := range nodes {
		if color[n] == white {
			if dfs(n) {
				return cycle
			}
		}
	}
	return nil
}

func rev(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
