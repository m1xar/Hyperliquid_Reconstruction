package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin"
	blofinclient "github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/service/reconstructor/helpers"
)

func main() {
	var (
		days   = flag.Int("days", 30, "days of history to reconstruct, 0 for everything available")
		sample = flag.Int("sample", 3, "how many rows to print per section")
		only   = flag.String("only", "", "run a single section: balance, open, closed, snapshots, transactions, fundings, candles")
		pair   = flag.String("pair", "BTC-USDT", "instrument for the candles section")
		raw    = flag.Bool("raw", false, "print untouched JSON of the account endpoints and exit")
		audit  = flag.Bool("audit", false, "print every fill and the episode it lands in, then exit")
	)
	flag.Parse()

	creds := blofinclient.Credentials{
		APIKey:     os.Getenv("BLOFIN_API_KEY"),
		Secret:     os.Getenv("BLOFIN_API_SECRET"),
		Passphrase: os.Getenv("BLOFIN_PASSPHRASE"),
	}
	if creds.APIKey == "" || creds.Secret == "" || creds.Passphrase == "" {
		fmt.Fprintln(os.Stderr, "set BLOFIN_API_KEY, BLOFIN_API_SECRET and BLOFIN_PASSPHRASE")
		os.Exit(1)
	}

	status := blofin.GetAuthStatus(creds.APIKey, creds.Secret, creds.Passphrase)
	fmt.Printf("auth: %s\n", status)
	if status != "ok" {
		os.Exit(1)
	}

	client := blofinclient.NewBaseClient()
	failed := false

	if *raw {
		blofinclient.AttachAuth(client, creds)
		dumpRaw(client)
		return
	}

	if *audit {
		blofinclient.AttachAuth(client, creds)
		if err := auditEpisodes(client, *days); err != nil {
			fmt.Fprintf(os.Stderr, "audit: FAILED: %v\n", err)
			os.Exit(1)
		}
		return
	}

	run := func(name string, fn func() error) {
		if *only != "" && *only != name {
			return
		}
		started := time.Now()
		if err := fn(); err != nil {
			failed = true
			fmt.Fprintf(os.Stderr, "%s: FAILED after %s: %v\n", name, since(started), err)
			return
		}
		fmt.Printf("%s: ok in %s\n\n", name, since(started))
	}

	run("balance", func() error {
		balance, err := blofin.GetCurrentBalance(client, creds)
		if err != nil {
			return err
		}
		fmt.Printf("  total equity: %.8f\n", *balance)
		return nil
	})

	run("open", func() error {
		positions, err := blofin.GetOpenPositions(client, creds)
		if err != nil {
			return err
		}
		fmt.Printf("  open positions: %d\n", len(positions))
		for i, pos := range positions {
			if i >= *sample {
				break
			}
			fmt.Printf("  %s %s amount=%.8f entry=%.8f mark=%.8f lev=%d orders=%d\n",
				pos.Side, pos.Pair, pos.Amount, pos.EntryPrice, pos.CurrentPrice, pos.Multiplier, len(pos.Orders))
		}
		return nil
	})

	run("closed", func() error {
		positions, err := blofin.GetBuiltPositions(client, creds, *days)
		if err != nil {
			return err
		}
		fmt.Printf("  closed positions: %d\n", len(positions))
		for i, pos := range positions {
			if i >= *sample {
				break
			}
			fmt.Printf("  %s\n", encode(pos))
		}
		for i, pos := range positions {
			if i >= *sample {
				break
			}
			fmt.Printf("  pnl check %s: gross=%.8f fee=%.8f funding=%.8f net=%.8f gross-fee+funding=%.8f\n",
				pos.Pair, pos.Pnl, pos.Commission, pos.Funding, pos.NetPnl,
				pos.Pnl-pos.Commission+pos.Funding)
		}
		return nil
	})

	run("snapshots", func() error {
		snapshots, err := blofin.GetBalanceSnapshots(client, creds, *days)
		if err != nil {
			return err
		}
		fmt.Printf("  snapshots: %d\n", len(snapshots))
		if len(snapshots) > 0 {
			first, last := snapshots[0], snapshots[len(snapshots)-1]
			fmt.Printf("  %s %.8f ... %s %.8f\n",
				first.CreatedAt.Format(time.RFC3339), first.Balance,
				last.CreatedAt.Format(time.RFC3339), last.Balance)
		}
		return nil
	})

	run("transactions", func() error {
		transactions, err := blofin.GetTransactions(client, creds, *days)
		if err != nil {
			return err
		}
		fmt.Printf("  transactions: %d\n", len(transactions))
		for i, tx := range transactions {
			if i >= *sample {
				break
			}
			fmt.Printf("  %s %s %.8f\n", tx.Time.Format(time.RFC3339), tx.Type, tx.Amount)
		}
		return nil
	})

	run("fundings", func() error {
		fundings, err := blofin.GetFundings(client, creds, *days)
		if err != nil {
			return err
		}
		fmt.Printf("  fundings: %d (BloFin keeps 7 days)\n", len(fundings))
		for i, f := range fundings {
			if i >= *sample {
				break
			}
			fmt.Printf("  %s %s %.8f\n", f.CreatedAt.Format(time.RFC3339), f.Pair, f.Amount)
		}
		return nil
	})

	run("candles", func() error {
		end := time.Now()
		candles, err := blofin.GetCandles(client, *pair, "1m", end.Add(-2*time.Hour), end)
		if err != nil {
			return err
		}
		fmt.Printf("  candles %s 1m over 2h: %d\n", *pair, len(candles))
		if len(candles) > 0 {
			c := candles[0]
			fmt.Printf("  newest ts=%s o=%s h=%s l=%s c=%s\n", c.Ts, c.O, c.H, c.L, c.C)
		}
		return nil
	})

	if failed {
		os.Exit(1)
	}
}

func auditEpisodes(client *resty.Client, days int) error {
	startMs := int64(0)
	if days > 0 {
		startMs = time.Now().AddDate(0, 0, -days).UnixMilli()
	}

	fills, err := executors.FetchAllFills(client, blofinclient.BaseURL, startMs)
	if err != nil {
		return err
	}

	instruments, err := executors.FetchInstruments(client, blofinclient.BaseURL)
	if err != nil {
		return err
	}

	sort.Slice(fills, func(i, j int) bool {
		return fills[i].Ts < fills[j].Ts
	})

	fmt.Printf("fills reported by the exchange: %d\n", len(fills))
	for _, f := range fills {
		instrument := executors.Instrument(instruments, f.InstID)
		fmt.Printf("  %s %-10s %-4s size=%s contracts (%.8f base) price=%s pnl=%s fee=%s order=%s\n",
			time.UnixMilli(toInt64(f.Ts)).UTC().Format("2006-01-02 15:04:05"),
			f.InstID, f.Side, f.FillSize,
			helpers.ContractsToBase(f.FillSize, instrument),
			f.FillPrice, f.FillPnl, f.Fee, f.OrderID)
	}

	episodes := helpers.BuildEpisodes(fills, instruments)
	fmt.Printf("\nepisodes cut from those fills: %d\n", len(episodes))
	for i, ep := range episodes {
		state := "OPEN"
		if ep.Closed {
			state = "CLOSED"
		}
		fmt.Printf("  #%d %s %s %s peak=%.8f opened=%s closed=%s fills=%d\n",
			i+1, state, ep.InstID, ep.Side(), ep.PeakSize,
			ep.OpenAt.UTC().Format("2006-01-02 15:04:05"),
			ep.CloseAt.UTC().Format("2006-01-02 15:04:05"),
			len(ep.Parts))
		for _, part := range ep.Parts {
			fmt.Printf("       %s %-4s %.8f @ %.8f pnl=%s order=%s\n",
				part.At.UTC().Format("2006-01-02 15:04:05"),
				part.Fill.Side, part.Size, part.Price, part.Fill.FillPnl, part.Fill.OrderID)
		}
	}

	return nil
}

func toInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func dumpRaw(client *resty.Client) {
	now := time.Now()
	begin90d := fmt.Sprint(now.AddDate(0, 0, -90).UnixMilli())
	begin7d := fmt.Sprint(now.AddDate(0, 0, -7).UnixMilli())
	nowMs := fmt.Sprint(now.UnixMilli())

	requests := []struct {
		path   string
		params map[string]string
	}{
		{path: "/api/v1/account/balance"},
		{path: "/api/v1/account/balance", params: map[string]string{"productType": "USDC-FUTURES"}},
		{path: "/api/v1/account/balance", params: map[string]string{"productType": "COIN-FUTURES"}},
		{path: "/api/v1/asset/balances", params: map[string]string{"accountType": "usdc_contract"}},
		{path: "/api/v1/account/positions"},
		{path: "/api/v1/account/positions-history", params: map[string]string{"limit": "5"}},
		{path: "/api/v1/account/positions-history", params: map[string]string{"limit": "5", "instId": "SOL-USDC"}},
		{path: "/api/v1/account/positions-history", params: map[string]string{"limit": "5", "productType": "USDC-FUTURES"}},
		{path: "/api/v1/account/positions-history", params: map[string]string{"limit": "5", "begin": begin90d, "end": nowMs}},
		{path: "/api/v1/account/funding-fees", params: map[string]string{"limit": "5", "begin": begin7d, "end": nowMs}},
		{path: "/api/v1/asset/bills", params: map[string]string{"limit": "10"}},
	}

	for _, r := range requests {
		body, err := blofinclient.DoGetRaw(client, blofinclient.BaseURL, r.path, r.params)
		fmt.Printf("=== GET %s %v\n", r.path, r.params)
		if err != nil {
			fmt.Printf("  error: %v\n", err)
		}
		fmt.Printf("%s\n\n", truncate(body, 4000))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "... (truncated)"
}

func encode(v any) string {
	encoded, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(encoded)
}

func since(started time.Time) time.Duration {
	return time.Since(started).Round(time.Millisecond)
}
