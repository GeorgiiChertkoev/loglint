package rules

// CheckNoSensitiveData reports if the message suggests logging sensitive data.
// hasVar is true when the log argument is a concatenation containing non-literal
// operands (variables, function calls). matchPattern tests the message against
// configured sensitive-data regexes.
func CheckNoSensitiveData(msg string, hasVar bool, matchPattern func(string) bool) string {
	if matchPattern == nil {
		return ""
	}
	if matchPattern(msg) {
		if hasVar {
			return "log message must not contain potentially sensitive data (avoid concatenating variables with sensitive keywords)"
		}
		return "log message must not contain potentially sensitive data (e.g. password, token, api_key)"
	}
	return ""
}
