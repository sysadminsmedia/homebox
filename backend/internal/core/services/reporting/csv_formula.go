package reporting

import "strings"

// Spreadsheet applications (Excel, LibreOffice, Google Sheets) evaluate a cell
// as a formula when it starts with one of these characters. Exported values are
// user-controlled, so they are neutralized to prevent CSV/formula injection
// (CWE-1236). See https://owasp.org/www-community/attacks/CSV_Injection
const csvFormulaTriggers = "=+-@\t\r"

// csvNeedsNeutralizing reports whether s would be interpreted as a formula, or
// already looks like a neutralized value (any number of leading quotes before a
// trigger character), so that every value survives an export/import round trip.
func csvNeedsNeutralizing(s string) bool {
	s = strings.TrimLeft(s, "'")
	return s != "" && strings.IndexByte(csvFormulaTriggers, s[0]) >= 0
}

// neutralizeCSVCell prefixes a single quote to values that a spreadsheet would
// otherwise evaluate as a formula. Only use it on text cells; numeric cells
// such as negative prices must stay as-is.
func neutralizeCSVCell(s string) string {
	if csvNeedsNeutralizing(s) {
		return "'" + s
	}
	return s
}

// restoreCSVCell reverses neutralizeCSVCell so that exported files can be
// re-imported without gaining a leading quote.
func restoreCSVCell(s string) string {
	if len(s) > 1 && s[0] == '\'' && csvNeedsNeutralizing(s[1:]) {
		return s[1:]
	}
	return s
}
