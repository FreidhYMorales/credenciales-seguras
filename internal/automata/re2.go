package automata

import "strings"

// ToRE2 translates the AST to an anchored RE2 expression, suitable for
// regexp.Compile. It is the artefact for steps 5 and 6 of the assignment: the
// translation is derived from the AST, not typed by hand.
func ToRE2(n Node) string {
	return `\A(?:` + ToRE2Body(n) + `)\z`
}

// ToRE2Body is the same translation without anchors, for display.
func ToRE2Body(n Node) string {
	switch t := n.(type) {
	case *leafNode:
		return "[" + re2Class(t.chars) + "]"
	case *epsNode:
		return "(?:)"
	case *catNode:
		return ToRE2Body(t.a) + ToRE2Body(t.b)
	case *altNode:
		return "(?:" + ToRE2Body(t.a) + "|" + ToRE2Body(t.b) + ")"
	case *starNode:
		return "(?:" + ToRE2Body(t.a) + ")*"
	}
	panic("automata: unknown node type")
}

// re2Class escapes the characters that are special inside a RE2 character
// class. The alphabet includes all of them, so a naive [...] would silently
// change the language.
func re2Class(chars []rune) string {
	var b strings.Builder
	b.Grow(len(chars) * 2)
	for _, r := range chars {
		switch r {
		case '\\', ']', '^', '-', '[':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
