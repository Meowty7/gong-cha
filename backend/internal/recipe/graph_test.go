package recipe

import (
	"reflect"
	"testing"
)

// TestDetectCycle_SelfCycle covers CP06: a product as a direct component of its own recipe.
func TestDetectCycle_SelfCycle(t *testing.T) {
	edges := []Edge{{From: "ST002", To: "ST002"}}
	cycle := DetectCycle(edges)
	if cycle == nil {
		t.Fatal("expected a self-cycle, got none")
	}
	if cycle[0] != "ST002" || cycle[len(cycle)-1] != "ST002" {
		t.Errorf("cycle should start and end at ST002, got %v", cycle)
	}
}

// TestDetectCycle_IndirectCycle covers CP06: an indirect cycle ST002 -> ST009 -> ST005 -> ST002.
func TestDetectCycle_IndirectCycle(t *testing.T) {
	edges := []Edge{
		{From: "ST002", To: "ST009"},
		{From: "ST009", To: "ST005"},
		{From: "ST005", To: "ST002"},
	}
	if cycle := DetectCycle(edges); cycle == nil {
		t.Fatal("expected an indirect cycle, got none")
	}
}

// TestDetectCycle_NoCycle accepts a valid chained recipe graph.
func TestDetectCycle_NoCycle(t *testing.T) {
	// PT001 -> ST002, ST009, ST001 (ST002 -> MP002, MP024; no cycle)
	edges := []Edge{
		{From: "PT001", To: "ST002"},
		{From: "PT001", To: "ST009"},
		{From: "ST002", To: "MP002"},
		{From: "ST009", To: "MP005"},
	}
	if cycle := DetectCycle(edges); cycle != nil {
		t.Fatalf("expected no cycle, got %v", cycle)
	}
}

// TestDetectCycle_EmptyGraph has no cycles.
func TestDetectCycle_EmptyGraph(t *testing.T) {
	if cycle := DetectCycle(nil); cycle != nil {
		t.Fatalf("expected no cycle on empty graph, got %v", cycle)
	}
}

// TestDetectCycle_DisconnectedComponents finds a cycle in one component only.
func TestDetectCycle_DisconnectedComponents(t *testing.T) {
	edges := []Edge{
		{From: "PT001", To: "ST002"},
		{From: "ST002", To: "MP002"}, // acyclic chain
		{From: "ST005", To: "ST005"}, // separate self-cycle
	}
	if cycle := DetectCycle(edges); cycle == nil {
		t.Fatal("expected to find the ST005 self-cycle in a disconnected graph")
	}
}

// TestDetectCycle_LargerChainedRecipes exercises the official-style chains.
func TestDetectCycle_LargerChainedRecipes(t *testing.T) {
	// PT010 -> ST010, ST005; ST005 -> MP009, MP024, ST001; ST001 -> MP001, MP024
	edges := []Edge{
		{From: "PT010", To: "ST010"},
		{From: "PT010", To: "ST005"},
		{From: "ST005", To: "MP009"},
		{From: "ST005", To: "MP024"},
		{From: "ST005", To: "ST001"},
		{From: "ST001", To: "MP001"},
		{From: "ST001", To: "MP024"},
		{From: "ST010", To: "MP020"},
		{From: "ST010", To: "MP024"},
	}
	if cycle := DetectCycle(edges); cycle != nil {
		t.Fatalf("expected no cycle in valid chains, got %v", cycle)
	}
}

// silence unused import
var _ = reflect.DeepEqual
