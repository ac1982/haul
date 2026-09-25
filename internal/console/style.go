// Package console is everything haul writes for people: styled status lines, terminal measurements and the
// arrow-key picker. Log lines go to stdout, or to stderr when stdout is reserved for data (--json).
package console

// Style wraps text in ANSI styles, or leaves it plain.
type Style struct{ Color bool }

// PlainStyle emits no escape codes.
var PlainStyle = Style{}

func (s Style) wrap(text, code string) string {
	if !s.Color {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func (s Style) Bold(t string) string     { return s.wrap(t, "1") }
func (s Style) Dim(t string) string      { return s.wrap(t, "2") }
func (s Style) Red(t string) string      { return s.wrap(t, "31") }
func (s Style) Green(t string) string    { return s.wrap(t, "32") }
func (s Style) Yellow(t string) string   { return s.wrap(t, "33") }
func (s Style) Cyan(t string) string     { return s.wrap(t, "36") }
func (s Style) Gray(t string) string     { return s.wrap(t, "90") }
func (s Style) BoldCyan(t string) string { return s.wrap(t, "1;36") }
