package automata

import "sort"

// treePos is one position of the augmented syntax tree (r)#.
type treePos struct {
	leaf *leafNode
}

type treeBuilder struct {
	pos    []treePos // 1-based: index 0 is a dummy
	follow [][]int   // follow[p], sorted
}

func (t *treeBuilder) addFollow(p int, first []int) {
	for len(t.follow) <= p {
		t.follow = append(t.follow, nil)
	}
	t.follow[p] = unionSorted(t.follow[p], first)
}

// walk returns nullable, first and last for n, filling follow as a side effect.
func (t *treeBuilder) walk(n Node) (nullable bool, first, last []int) {
	switch x := n.(type) {
	case *leafNode:
		p := len(t.pos)
		t.pos = append(t.pos, treePos{leaf: x})
		return false, []int{p}, []int{p}
	case *epsNode:
		return true, nil, nil
	case *starNode:
		_, f, l := t.walk(x.a)
		for _, p := range l {
			t.addFollow(p, f)
		}
		return true, f, l
	case *catNode:
		n1, f1, l1 := t.walk(x.a)
		n2, f2, l2 := t.walk(x.b)
		for _, p := range l1 {
			t.addFollow(p, f2)
		}
		first = f1
		if n1 {
			first = unionSorted(f1, f2)
		}
		last = l2
		if n2 {
			last = unionSorted(l1, l2)
		}
		return n1 && n2, first, last
	case *altNode:
		n1, f1, l1 := t.walk(x.a)
		n2, f2, l2 := t.walk(x.b)
		return n1 || n2, unionSorted(f1, f2), unionSorted(l1, l2)
	}
	panic("automata: unknown node type")
}

// TreeDFA builds the DFA with the syntax-tree (followpos) method.
func TreeDFA(root Node, atoms []Atom, c2a map[rune]int) *DFA {
	t := &treeBuilder{pos: []treePos{{}}}
	endLeaf := Leaf([]rune{'#'}, "#").(*leafNode)
	nullable, first, _ := t.walk(Cat(root, endLeaf))
	_ = nullable
	end := len(t.pos) - 1

	b := newDFABuilder(atoms, c2a)
	start, _ := b.intern(first)
	queue := []int{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		set := b.sets[cur]
		for ai, at := range atoms {
			rep := at.Rep()
			var next []int
			for _, p := range set {
				if p == end {
					continue
				}
				if contains(t.pos[p].leaf.chars, rep) {
					next = unionSorted(next, t.follow[p])
				}
			}
			if len(next) == 0 {
				continue
			}
			target, isNew := b.intern(next)
			b.delta[cur][ai] = target
			if isNew {
				queue = append(queue, target)
			}
		}
	}

	return b.finish(start, func(st []int) bool {
		i := sort.SearchInts(st, end)
		return i < len(st) && st[i] == end
	})
}

// unionSorted merges two sorted int slices without duplicates.
func unionSorted(a, b []int) []int {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	out := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			out = append(out, a[i])
			i++
		case a[i] > b[j]:
			out = append(out, b[j])
			j++
		default:
			out = append(out, a[i])
			i++
			j++
		}
	}
	out = append(out, a[i:]...)
	out = append(out, b[j:]...)
	return out
}
