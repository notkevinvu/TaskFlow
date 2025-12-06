package graph

// CycleDetector checks for cycles in a directed graph using DFS
// Used to prevent circular dependencies between tasks
type CycleDetector struct {
	// adjacency maps each node to its outgoing edges (node -> nodes it points to)
	// For dependencies: task_id -> [blocked_by_ids]
	adjacency map[string][]string
}

// Color represents node visit state during DFS traversal
type color int

const (
	white color = iota // Not visited
	gray               // Currently being visited (in recursion stack)
	black              // Finished visiting
)

// NewCycleDetector creates a new detector with the given edges
// edges is a map of node -> list of nodes it depends on
func NewCycleDetector(edges map[string][]string) *CycleDetector {
	// Copy the map to avoid external mutation
	adj := make(map[string][]string, len(edges))
	for k, v := range edges {
		adj[k] = append([]string{}, v...)
	}
	return &CycleDetector{adjacency: adj}
}

// WouldCreateCycle checks if adding an edge from -> to would create a cycle
// Returns true if adding this edge would create a cycle
func (cd *CycleDetector) WouldCreateCycle(from, to string) bool {
	// Self-loop is always a cycle
	if from == to {
		return true
	}

	// Temporarily add the edge
	cd.adjacency[from] = append(cd.adjacency[from], to)
	defer func() {
		// Remove the temporary edge
		edges := cd.adjacency[from]
		cd.adjacency[from] = edges[:len(edges)-1]
	}()

	// Check if we can reach 'from' starting from 'to'
	// If we can, then adding from->to creates a cycle
	return cd.canReach(to, from, make(map[string]bool))
}

// canReach checks if we can reach 'target' starting from 'start'
// Uses DFS with visited tracking
func (cd *CycleDetector) canReach(start, target string, visited map[string]bool) bool {
	if start == target {
		return true
	}
	if visited[start] {
		return false
	}
	visited[start] = true

	for _, neighbor := range cd.adjacency[start] {
		if cd.canReach(neighbor, target, visited) {
			return true
		}
	}
	return false
}

// HasCycle checks if the current graph has any cycles
// Uses standard DFS coloring algorithm
func (cd *CycleDetector) HasCycle() bool {
	colors := make(map[string]color)

	// Initialize all nodes as white (unvisited)
	for node := range cd.adjacency {
		colors[node] = white
	}

	// Run DFS from each unvisited node
	for node := range cd.adjacency {
		if colors[node] == white {
			if cd.hasCycleDFS(node, colors) {
				return true
			}
		}
	}
	return false
}

// hasCycleDFS performs DFS and returns true if a cycle is found
func (cd *CycleDetector) hasCycleDFS(node string, colors map[string]color) bool {
	colors[node] = gray // Mark as being visited

	for _, neighbor := range cd.adjacency[node] {
		// Initialize neighbor color if not in map
		if _, exists := colors[neighbor]; !exists {
			colors[neighbor] = white
		}

		if colors[neighbor] == gray {
			// Found a back edge - cycle detected
			return true
		}
		if colors[neighbor] == white {
			if cd.hasCycleDFS(neighbor, colors) {
				return true
			}
		}
	}

	colors[node] = black // Mark as finished
	return false
}

// GetAllReachable returns all nodes reachable from the given start node
// Useful for finding all transitive dependencies
func (cd *CycleDetector) GetAllReachable(start string) []string {
	visited := make(map[string]bool)
	var result []string

	cd.collectReachable(start, visited, &result)
	return result
}

// collectReachable recursively collects all reachable nodes
func (cd *CycleDetector) collectReachable(node string, visited map[string]bool, result *[]string) {
	for _, neighbor := range cd.adjacency[node] {
		if !visited[neighbor] {
			visited[neighbor] = true
			*result = append(*result, neighbor)
			cd.collectReachable(neighbor, visited, result)
		}
	}
}
