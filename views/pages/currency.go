package pages

import "fmt"

type currencySpec struct {
	symbol         string
	thousandsSep   byte
	symbolIsPrefix bool
	locale         string
}

func currencyFormatSpec(currencyCode string) currencySpec {
	switch currencyCode {
	case "IDR":
		return currencySpec{symbol: "Rp", thousandsSep: '.', symbolIsPrefix: true, locale: "id-ID"}
	case "EUR":
		return currencySpec{symbol: "€", thousandsSep: '.', symbolIsPrefix: false, locale: "de-DE"}
	case "SGD":
		return currencySpec{symbol: "S$", thousandsSep: ',', symbolIsPrefix: true, locale: "en-SG"}
	case "GBP":
		return currencySpec{symbol: "£", thousandsSep: ',', symbolIsPrefix: true, locale: "en-GB"}
	case "JPY":
		return currencySpec{symbol: "¥", thousandsSep: ',', symbolIsPrefix: true, locale: "ja-JP"}
	default:
		return currencySpec{symbol: "$", thousandsSep: ',', symbolIsPrefix: true, locale: "en-US"}
	}
}

func currencySymbol(currencyCode string) string {
	return currencyFormatSpec(currencyCode).symbol
}

func currencyLocale(currencyCode string) string {
	return currencyFormatSpec(currencyCode).locale
}

func currencySymbolIsPrefix(currencyCode string) bool {
	return currencyFormatSpec(currencyCode).symbolIsPrefix
}

func formatMoney(amount int, currencyCode string) string {
	spec := currencyFormatSpec(currencyCode)
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	formatted := formatIntegerWithSeparator(amount, spec.thousandsSep)
	if spec.symbolIsPrefix {
		return fmt.Sprintf("%s%s %s", sign, spec.symbol, formatted)
	}
	return fmt.Sprintf("%s%s %s", sign, formatted, spec.symbol)
}

func formatIntegerWithSeparator(n int, sep byte) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	first := len(s) % 3
	if first == 0 {
		first = 3
	}

	out := make([]byte, 0, len(s)+(len(s)-1)/3)
	out = append(out, s[:first]...)
	for i := first; i < len(s); i += 3 {
		out = append(out, sep)
		out = append(out, s[i:i+3]...)
	}
	return string(out)
}
