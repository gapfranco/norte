package main

import (
	"strings"
	"unicode"
)

// normalizarCNPJ removes punctuation/spaces and uppercases alphanumeric chars.
// Empty/whitespace input returns "".
func normalizarCNPJ(cnpj string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}
		return -1
	}, strings.TrimSpace(cnpj))
}

// validarCNPJ validates Brazilian CNPJ in both the legacy all-numeric format
// and the new alphanumeric format (IN RFB 2.229/2024, effective 2026).
// Accepts formatted (XX.XXX.XXX/XXXX-DD) or stripped strings.
// An empty string is considered valid (field is optional).
func validarCNPJ(cnpj string) bool {
	cnpj = normalizarCNPJ(cnpj)
	if cnpj == "" {
		return true
	}

	if len(cnpj) != 14 {
		return false
	}

	// Check digits (positions 13-14) must be numeric
	if cnpj[12] < '0' || cnpj[12] > '9' || cnpj[13] < '0' || cnpj[13] > '9' {
		return false
	}

	// Reject sequences of a single repeated character
	allSame := true
	for i := 1; i < 14; i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// Mapping per IN RFB 2.229/2024: ord(c) - ord('0') for all alphanumeric chars.
	// Digits 0-9 → 0-9; letters A-Z → 17-42.
	charVal := func(c byte) int {
		return int(c - '0')
	}

	calcDigit := func(s string, weights []int) int {
		sum := 0
		for i, w := range weights {
			sum += charVal(s[i]) * w
		}
		if r := sum % 11; r >= 2 {
			return 11 - r
		}
		return 0
	}

	w1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	w2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	return calcDigit(cnpj, w1) == int(cnpj[12]-'0') &&
		calcDigit(cnpj, w2) == int(cnpj[13]-'0')
}
