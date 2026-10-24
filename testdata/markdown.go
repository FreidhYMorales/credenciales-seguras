package testdata

import (
	"fmt"
	"strings"

	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
)

// Markdown renders the graded pairs as the two tables of the final document:
// the ten generated pairs with their seeds, and the ten authored pairs with the
// rule that rejects the defective half.
//
// The output is a fragment, not a document: it starts at the subsection level so
// it nests under the heading of whichever chapter includes it. Passwords are
// written as code spans, because they contain pipes, backticks and asterisks
// that a bare table cell would swallow.
func Markdown(pairs []Pair) []byte {
	var b strings.Builder

	b.WriteString("### Pares válidos\n\n")
	b.WriteString("Los diez pares válidos los generó la herramienta caminando el autómata producto. ")
	b.WriteString("Las semillas quedan registradas, así que `make pairs` los reproduce byte a byte.\n\n")
	b.WriteString("| # | Correo | Contraseña | Semillas (correo / contraseña) |\n")
	b.WriteString("|---:|---|---|---|\n")
	for _, p := range pairs {
		if !p.EmailEsperado || !p.PasswordEsperado {
			continue
		}
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			p.ID, codeSpan(p.Email), codeSpan(p.Password), semillas(p)))
	}

	b.WriteString("\n### Pares inválidos\n\n")
	b.WriteString("Cada fila tiene exactamente una credencial defectuosa y nombra la regla que la rechaza; ")
	b.WriteString("la otra mitad del par es válida.\n\n")
	b.WriteString("| # | Correo | Contraseña | Qué falla |\n")
	b.WriteString("|---:|---|---|---|\n")
	for _, p := range pairs {
		if p.EmailEsperado && p.PasswordEsperado {
			continue
		}
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			p.ID, codeSpan(p.Email), codeSpan(p.Password), falla(p)))
	}

	return []byte(b.String())
}

// semillas prints the two recorded seeds, or a dash when the pair was authored
// by hand.
func semillas(p Pair) string {
	if p.SeedEmail == nil || p.SeedPassword == nil {
		return "—"
	}
	return fmt.Sprintf("%d / %d", *p.SeedEmail, *p.SeedPassword)
}

// falla names the half that fails and the rule that rejects it, e.g.
// "correo · **C2** Parte local".
func falla(p Pair) string {
	lado := "contraseña"
	if !p.EmailEsperado {
		lado = "correo"
	}
	if p.ReglaFallida == nil {
		return lado
	}
	id := *p.ReglaFallida
	if r := rules.Get(id); r != nil {
		return fmt.Sprintf("%s · **%s** %s", lado, r.ID, r.Nombre)
	}
	return fmt.Sprintf("%s · **%s**", lado, id)
}

// codeSpan wraps s in the shortest backtick fence that survives its content, so
// a password containing a backtick still renders as literal text. The pipe is
// escaped because a table cell treats it as a column separator even inside a
// code span.
func codeSpan(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)

	run, max := 0, 0
	for _, r := range s {
		if r == '`' {
			run++
			if run > max {
				max = run
			}
			continue
		}
		run = 0
	}

	fence := strings.Repeat("`", max+1)
	pad := ""
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		pad = " "
	}
	return fence + pad + s + pad + fence
}
