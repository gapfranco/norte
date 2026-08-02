package main

import (
	"net/mail"
	"strings"
)

// normalizeEmail trims and lowercases an e-mail used as login (users.usuario).
func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// isValidEmail reports whether s is a bare e-mail address suitable for users.usuario.
func isValidEmail(s string) bool {
	s = normalizeEmail(s)
	if s == "" || strings.ContainsAny(s, " <>") {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	return strings.EqualFold(addr.Address, s) && strings.Count(s, "@") == 1
}
