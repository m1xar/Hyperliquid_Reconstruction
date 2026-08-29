package helpers

import (
	"sort"
	"strings"

	"github.com/m1xar/scope360-reconstruction/pkg/kraken/connector/kraken/models"
)

func RawSymbolByPair(symbols []string, pairBySymbol map[string]string) map[string]string {
	out := make(map[string]string, len(symbols))
	for _, symbol := range symbols {
		sym := strings.ToUpper(strings.TrimSpace(symbol))
		if sym == "" {
			continue
		}
		pair := NormalizePair(sym, pairBySymbol)
		if _, ok := out[pair]; !ok {
			out[pair] = sym
		}
	}
	return out
}

func SymbolsFromFillsAndEvents(fills []models.Fill, events []models.PositionEventElement) []string {
	set := make(map[string]struct{})
	for _, fill := range fills {
		set[strings.ToUpper(fill.Symbol)] = struct{}{}
	}
	for _, ev := range events {
		if ev.Event.PositionUpdate.Tradeable != "" {
			set[strings.ToUpper(ev.Event.PositionUpdate.Tradeable)] = struct{}{}
		}
	}
	return sortedKeys(set)
}

func SymbolsFromAccountLogs(logs []models.AccountLog) []string {
	set := make(map[string]struct{})
	for _, row := range logs {
		if row.Contract != "" {
			set[strings.ToUpper(row.Contract)] = struct{}{}
		}
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func SymbolFromPair(pair string) string {
	p := strings.ToUpper(strings.TrimSpace(pair))
	if strings.HasPrefix(p, "PF_") || strings.HasPrefix(p, "PI_") || strings.HasPrefix(p, "FI_") || strings.HasPrefix(p, "FF_") {
		return p
	}
	return "PF_" + p
}
