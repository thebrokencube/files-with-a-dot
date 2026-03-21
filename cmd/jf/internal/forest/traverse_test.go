package forest

import (
	"testing"
)

func buildTestTree() []*Node {
	// A (root)
	//   B
	//     D
	//     E
	//   C
	e := &Node{Key: "BEN-5", Label: "E", File: "a/b/e.md"}
	d := &Node{Key: "BEN-4", Label: "D", File: "a/b/d.md"}
	c := &Node{Key: "BEN-3", Label: "C", File: "a/c.md"}
	b := &Node{Key: "BEN-2", Label: "B", File: "a/b/README.md", Children: []*Node{d, e}}
	a := &Node{Key: "BEN-1", Label: "A", File: "a/README.md", Children: []*Node{b, c}}

	d.Parent = b
	e.Parent = b
	b.Parent = a
	c.Parent = a

	return []*Node{a}
}

func keys(nodes []*Node) []string {
	var result []string
	for _, n := range nodes {
		result = append(result, n.Key)
	}
	return result
}

func TestPostOrder(t *testing.T) {
	roots := buildTestTree()
	result := PostOrder(roots)
	got := keys(result)
	// Post-order: D, E, B, C, A
	expected := []string{"BEN-4", "BEN-5", "BEN-2", "BEN-3", "BEN-1"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d nodes, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], got[i])
		}
	}
}

func TestPreOrder(t *testing.T) {
	roots := buildTestTree()
	result := PreOrder(roots)
	got := keys(result)
	// Pre-order: A, B, D, E, C
	expected := []string{"BEN-1", "BEN-2", "BEN-4", "BEN-5", "BEN-3"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d nodes, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], got[i])
		}
	}
}

func TestMultiRootTraversal(t *testing.T) {
	r1 := &Node{Key: "BEN-1", Label: "R1", File: "r1/README.md"}
	r2 := &Node{Key: "BEN-2", Label: "R2", File: "r2/README.md",
		Children: []*Node{{Key: "BEN-3", Label: "Child", File: "r2/child.md"}}}
	r2.Children[0].Parent = r2

	roots := []*Node{r1, r2}

	post := keys(PostOrder(roots))
	// R1, then BEN-3, R2
	expected := []string{"BEN-1", "BEN-3", "BEN-2"}
	for i := range expected {
		if post[i] != expected[i] {
			t.Errorf("post-order position %d: expected %s, got %s", i, expected[i], post[i])
		}
	}

	pre := keys(PreOrder(roots))
	// R1, R2, BEN-3
	expectedPre := []string{"BEN-1", "BEN-2", "BEN-3"}
	for i := range expectedPre {
		if pre[i] != expectedPre[i] {
			t.Errorf("pre-order position %d: expected %s, got %s", i, expectedPre[i], pre[i])
		}
	}
}

func TestSubtree(t *testing.T) {
	roots := buildTestTree()
	node, err := Subtree(roots, "BEN-2")
	if err != nil {
		t.Fatal(err)
	}
	if node.Key != "BEN-2" {
		t.Errorf("expected BEN-2, got %s", node.Key)
	}
	// Subtree of B includes D and E
	sub := PostOrder([]*Node{node})
	got := keys(sub)
	expected := []string{"BEN-4", "BEN-5", "BEN-2"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d nodes in subtree, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("subtree position %d: expected %s, got %s", i, expected[i], got[i])
		}
	}
}

func TestResolveByKey(t *testing.T) {
	roots := buildTestTree()
	node, err := Resolve(roots, "BEN-4")
	if err != nil {
		t.Fatal(err)
	}
	if node.Label != "D" {
		t.Errorf("expected D, got %s", node.Label)
	}
}

func TestResolveByStem(t *testing.T) {
	roots := buildTestTree()
	// e.md -> stem "e"
	node, err := Resolve(roots, "e")
	if err != nil {
		t.Fatal(err)
	}
	if node.Key != "BEN-5" {
		t.Errorf("expected BEN-5, got %s", node.Key)
	}
}

func TestResolveByKeyIsCaseInsensitive(t *testing.T) {
	roots := buildTestTree()
	node, err := Resolve(roots, "ben-1")
	if err != nil {
		t.Fatal(err)
	}
	if node.Key != "BEN-1" {
		t.Errorf("expected BEN-1, got %s", node.Key)
	}
}

func TestResolveNotFound(t *testing.T) {
	roots := buildTestTree()
	_, err := Resolve(roots, "NOPE-999")
	if err == nil {
		t.Error("expected error for missing node")
	}
}

func TestResolveByFilePath(t *testing.T) {
	roots := buildTestTree()
	node, err := Resolve(roots, "a/b/d.md")
	if err != nil {
		t.Fatal(err)
	}
	if node.Key != "BEN-4" {
		t.Errorf("expected BEN-4, got %s", node.Key)
	}
}
