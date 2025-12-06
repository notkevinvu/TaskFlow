package graph

import (
	"testing"
)

func TestNewCycleDetector(t *testing.T) {
	edges := map[string][]string{
		"A": {"B", "C"},
		"B": {"C"},
	}

	cd := NewCycleDetector(edges)

	// Verify the detector was created and edges were copied
	if cd == nil {
		t.Fatal("Expected non-nil CycleDetector")
	}

	// Modify original map - should not affect detector
	edges["A"] = append(edges["A"], "D")
	if len(cd.adjacency["A"]) != 2 {
		t.Error("CycleDetector should have its own copy of edges")
	}
}

func TestWouldCreateCycle_SelfLoop(t *testing.T) {
	cd := NewCycleDetector(map[string][]string{})

	if !cd.WouldCreateCycle("A", "A") {
		t.Error("Self-loop should create a cycle")
	}
}

func TestWouldCreateCycle_DirectCycle(t *testing.T) {
	// A -> B exists
	edges := map[string][]string{
		"A": {"B"},
	}
	cd := NewCycleDetector(edges)

	// Adding B -> A would create A -> B -> A
	if !cd.WouldCreateCycle("B", "A") {
		t.Error("B -> A should create a cycle with existing A -> B")
	}
}

func TestWouldCreateCycle_IndirectCycle(t *testing.T) {
	// A -> B -> C exists
	edges := map[string][]string{
		"A": {"B"},
		"B": {"C"},
	}
	cd := NewCycleDetector(edges)

	// Adding C -> A would create A -> B -> C -> A
	if !cd.WouldCreateCycle("C", "A") {
		t.Error("C -> A should create a cycle with existing A -> B -> C")
	}
}

func TestWouldCreateCycle_NoCycle(t *testing.T) {
	// A -> B exists
	edges := map[string][]string{
		"A": {"B"},
	}
	cd := NewCycleDetector(edges)

	// Adding B -> C should not create a cycle
	if cd.WouldCreateCycle("B", "C") {
		t.Error("B -> C should not create a cycle")
	}

	// Adding A -> C should not create a cycle
	if cd.WouldCreateCycle("A", "C") {
		t.Error("A -> C should not create a cycle")
	}
}

func TestWouldCreateCycle_LongChain(t *testing.T) {
	// A -> B -> C -> D -> E exists
	edges := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"D"},
		"D": {"E"},
	}
	cd := NewCycleDetector(edges)

	// Adding E -> A would create a long cycle
	if !cd.WouldCreateCycle("E", "A") {
		t.Error("E -> A should create a cycle")
	}

	// Adding E -> B would also create a cycle
	if !cd.WouldCreateCycle("E", "B") {
		t.Error("E -> B should create a cycle")
	}

	// Adding E -> F should not create a cycle
	if cd.WouldCreateCycle("E", "F") {
		t.Error("E -> F should not create a cycle")
	}
}

func TestWouldCreateCycle_DiamondShape(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	edges := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
	}
	cd := NewCycleDetector(edges)

	// No cycle should exist
	if cd.HasCycle() {
		t.Error("Diamond shape should not have a cycle")
	}

	// Adding D -> A would create a cycle
	if !cd.WouldCreateCycle("D", "A") {
		t.Error("D -> A should create a cycle in diamond")
	}

	// Adding D -> B would create a cycle
	if !cd.WouldCreateCycle("D", "B") {
		t.Error("D -> B should create a cycle in diamond")
	}
}

func TestWouldCreateCycle_PreservesState(t *testing.T) {
	edges := map[string][]string{
		"A": {"B"},
	}
	cd := NewCycleDetector(edges)

	// Check cycle - should not modify state
	cd.WouldCreateCycle("B", "A")

	// Original edges should be preserved
	if len(cd.adjacency["A"]) != 1 || cd.adjacency["A"][0] != "B" {
		t.Error("WouldCreateCycle should not modify original edges")
	}

	// Edge from B should not exist
	if len(cd.adjacency["B"]) != 0 {
		t.Error("Temporary edge should be removed after check")
	}
}

func TestHasCycle_EmptyGraph(t *testing.T) {
	cd := NewCycleDetector(map[string][]string{})

	if cd.HasCycle() {
		t.Error("Empty graph should not have a cycle")
	}
}

func TestHasCycle_SingleNode(t *testing.T) {
	edges := map[string][]string{
		"A": {},
	}
	cd := NewCycleDetector(edges)

	if cd.HasCycle() {
		t.Error("Single node without self-loop should not have a cycle")
	}
}

func TestHasCycle_WithCycle(t *testing.T) {
	// A -> B -> C -> A
	edges := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	}
	cd := NewCycleDetector(edges)

	if !cd.HasCycle() {
		t.Error("Graph with A -> B -> C -> A should have a cycle")
	}
}

func TestHasCycle_DisconnectedWithCycle(t *testing.T) {
	// Disconnected: A -> B (no cycle), C -> D -> C (cycle)
	edges := map[string][]string{
		"A": {"B"},
		"C": {"D"},
		"D": {"C"},
	}
	cd := NewCycleDetector(edges)

	if !cd.HasCycle() {
		t.Error("Graph with disconnected cycle should detect the cycle")
	}
}

func TestGetAllReachable_Empty(t *testing.T) {
	cd := NewCycleDetector(map[string][]string{})

	result := cd.GetAllReachable("A")
	if len(result) != 0 {
		t.Error("Empty graph should return no reachable nodes")
	}
}

func TestGetAllReachable_Chain(t *testing.T) {
	// A -> B -> C -> D
	edges := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"D"},
	}
	cd := NewCycleDetector(edges)

	result := cd.GetAllReachable("A")
	if len(result) != 3 {
		t.Errorf("Expected 3 reachable nodes from A, got %d", len(result))
	}

	// Check all expected nodes are present
	found := make(map[string]bool)
	for _, node := range result {
		found[node] = true
	}
	for _, expected := range []string{"B", "C", "D"} {
		if !found[expected] {
			t.Errorf("Expected %s to be reachable from A", expected)
		}
	}
}

func TestGetAllReachable_Diamond(t *testing.T) {
	// A -> B, A -> C, B -> D, C -> D
	edges := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
	}
	cd := NewCycleDetector(edges)

	result := cd.GetAllReachable("A")
	if len(result) != 3 {
		t.Errorf("Expected 3 reachable nodes from A (B, C, D), got %d", len(result))
	}
}

func TestGetAllReachable_NoDuplicates(t *testing.T) {
	// Multiple paths to same node
	edges := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
		"D": {"E"},
	}
	cd := NewCycleDetector(edges)

	result := cd.GetAllReachable("A")
	seen := make(map[string]bool)
	for _, node := range result {
		if seen[node] {
			t.Errorf("Duplicate node %s in result", node)
		}
		seen[node] = true
	}
}

// Benchmark tests
func BenchmarkWouldCreateCycle_SmallGraph(b *testing.B) {
	edges := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"D"},
	}
	cd := NewCycleDetector(edges)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd.WouldCreateCycle("D", "A")
	}
}

func BenchmarkWouldCreateCycle_LargeGraph(b *testing.B) {
	// Create a chain of 100 nodes
	edges := make(map[string][]string)
	for i := 0; i < 99; i++ {
		from := string(rune('A' + i%26)) + string(rune('0'+i/26))
		to := string(rune('A' + (i+1)%26)) + string(rune('0'+(i+1)/26))
		edges[from] = []string{to}
	}
	cd := NewCycleDetector(edges)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd.WouldCreateCycle("Z3", "A0")
	}
}
