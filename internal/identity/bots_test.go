package identity

import (
	"strings"
	"testing"
)

func TestParseBotSpecCategoriesNamesAliases(t *testing.T) {
	filter, err := ParseBotSpec([]string{"ai", "crawler", "GPTBot"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filter.Categories[CategoryAICrawler] || !filter.Categories[CategoryAIAssistant] {
		t.Fatalf("alias 'ai' should enable both AI categories: %+v", filter.Categories)
	}
	if !filter.Categories[CategoryCrawler] {
		t.Fatalf("'crawler' category missing")
	}
	if !filter.Names["gptbot"] {
		t.Fatalf("exact name 'GPTBot' not registered: %+v", filter.Names)
	}
}

func TestParseBotSpecAllIsEmptyFilter(t *testing.T) {
	filter, err := ParseBotSpec([]string{"all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filter.Empty() {
		t.Fatalf("'all' should produce an empty (match-all) filter")
	}
}

func TestParseBotSpecUnknownTokenErrors(t *testing.T) {
	_, err := ParseBotSpec([]string{"notabot"})
	if err == nil {
		t.Fatal("expected error for unknown token")
	}
	if !strings.Contains(err.Error(), "notabot") {
		t.Fatalf("error should name the bad token: %v", err)
	}
}

func TestBotPoolFilterByName(t *testing.T) {
	filter, err := ParseBotSpec([]string{"GPTBot"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pool := BotPool(filter)
	if len(pool) != 1 {
		t.Fatalf("pool size = %d, want 1", len(pool))
	}
	if !strings.Contains(pool[0].Value, "GPTBot") {
		t.Fatalf("pool UA = %q, want a GPTBot UA", pool[0].Value)
	}
}

func TestBotPoolEmptyFilterCoversWholeCatalog(t *testing.T) {
	if got := len(BotPool(BotFilter{})); got != len(Bots) {
		t.Fatalf("empty filter pool = %d, want %d", got, len(Bots))
	}
}

func TestGeneratorUseBotsRestrictsPool(t *testing.T) {
	g := New(7, 64)
	filter, err := ParseBotSpec([]string{"ClaudeBot"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	g.UseBots(filter)
	// unknownRatio 100 => every request is a bot; the pool has one entry.
	for i := 0; i < 50; i++ {
		ident := g.Next(0, 100, 1)
		if !strings.Contains(ident.UserAgent, "ClaudeBot") {
			t.Fatalf("got UA %q, want ClaudeBot", ident.UserAgent)
		}
	}
}

func TestEveryCatalogBotHasUAAndCategory(t *testing.T) {
	valid := map[BotCategory]bool{}
	for _, c := range AllCategories {
		valid[c] = true
	}
	for _, bot := range Bots {
		if strings.TrimSpace(bot.UA) == "" {
			t.Fatalf("bot %q has empty UA", bot.Name)
		}
		if !valid[bot.Category] {
			t.Fatalf("bot %q has unknown category %q", bot.Name, bot.Category)
		}
		if bot.Weight <= 0 {
			t.Fatalf("bot %q has non-positive weight", bot.Name)
		}
	}
}

func TestBotPoolNoMatchFallsBackToAll(t *testing.T) {
	// A filter that matches nothing (constructed directly) should not stall.
	filter := BotFilter{Names: map[string]bool{"does-not-exist": true}}
	if got := len(BotPool(filter)); got != len(Bots) {
		t.Fatalf("fallback pool = %d, want full catalog %d", got, len(Bots))
	}
}
