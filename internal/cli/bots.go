package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zeb-link/hitmaker/v2/internal/identity"
	"github.com/zeb-link/hitmaker/v2/internal/ui/theme"
)

// botCatalogEntry is the machine-readable shape emitted by `bots --json`.
type botCatalogEntry struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Type     string `json:"botType"`
	UA       string `json:"userAgent"`
}

type botCategoryEntry struct {
	Category    string            `json:"category"`
	Description string            `json:"description"`
	Aliases     []string          `json:"aliases"`
	Bots        []botCatalogEntry `json:"bots"`
}

type botCatalog struct {
	Usage      string             `json:"usage"`
	Categories []botCategoryEntry `json:"categories"`
}

func newBotsCommand(root *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "bots",
		Short: "List every bot / AI-crawler identity Hitmaker can send",
		Long: `List the catalog of bot and AI-crawler identities Hitmaker can impersonate.

Each identity is sent as that agent's real, publicly documented User-Agent
string, so analytics that classify traffic by bot type will categorize it. Use
these values with --bots to shape traffic:

  hitmaker run --bots ai --bot-ratio 100 https://example.com/a   # only AI bots
  hitmaker run --bots GPTBot,ClaudeBot --bot-ratio 100 URL       # specific names
  hitmaker run --bots crawler --bot-ratio 40 URL                 # 40% search spiders
  hitmaker probe --bots PerplexityBot URL                        # single diagnostic hit

--bots accepts category values (ai_crawler, ai_assistant, crawler, fetcher, cli,
library), the alias "ai" (both AI categories), "all", or exact bot names.
--bot-ratio sets the percent of traffic that is bots (0-100).

Add --json for a machine-readable catalog.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if root.JSON {
				return writeJSON(buildBotCatalog())
			}
			printBotCatalog()
			return nil
		},
	}
}

func buildBotCatalog() botCatalog {
	grouped := identity.BotsByCategory()
	cat := botCatalog{
		Usage:      "hitmaker run --bots <category|name|ai|all> --bot-ratio <0-100> <url>",
		Categories: make([]botCategoryEntry, 0, len(identity.AllCategories)),
	}
	for _, category := range identity.AllCategories {
		bots := grouped[category]
		sort.Slice(bots, func(i, j int) bool { return bots[i].Name < bots[j].Name })
		entry := botCategoryEntry{
			Category:    string(category),
			Description: identity.CategoryDescription(category),
			Aliases:     categoryAliasesFor(category),
			Bots:        make([]botCatalogEntry, 0, len(bots)),
		}
		for _, bot := range bots {
			entry.Bots = append(entry.Bots, botCatalogEntry{
				Name:     bot.Name,
				Category: string(bot.Category),
				Type:     string(bot.Category),
				UA:       bot.UA,
			})
		}
		cat.Categories = append(cat.Categories, entry)
	}
	return cat
}

func printBotCatalog() {
	grouped := identity.BotsByCategory()
	fmt.Println(theme.Logo.Render("HITMAKER bots"))
	fmt.Printf("%s\n\n", theme.Subtle.Render("Use with:  hitmaker run --bots <category|name|ai|all> --bot-ratio <0-100> <url>"))
	total := 0
	for _, category := range identity.AllCategories {
		bots := grouped[category]
		total += len(bots)
		sort.Slice(bots, func(i, j int) bool { return bots[i].Name < bots[j].Name })
		aliases := categoryAliasesFor(category)
		fmt.Printf("%s  %s\n", theme.Focus.Render(string(category)), theme.Subtle.Render("aliases: "+joinComma(aliases)))
		fmt.Printf("  %s\n", theme.Subtle.Render(identity.CategoryDescription(category)))
		for _, bot := range bots {
			fmt.Printf("    %s\n", bot.Name)
		}
		fmt.Println()
	}
	fmt.Printf("%s %d identities across %d categories. Alias \"ai\" = ai_crawler + ai_assistant.\n",
		theme.Good.Render("•"), total, len(identity.AllCategories))
}

func categoryAliasesFor(c identity.BotCategory) []string {
	switch c {
	case identity.CategoryAICrawler:
		return []string{"ai_crawler", "ai"}
	case identity.CategoryAIAssistant:
		return []string{"ai_assistant", "assistant", "ai"}
	case identity.CategoryCrawler:
		return []string{"crawler", "search"}
	case identity.CategoryFetcher:
		return []string{"fetcher", "social", "preview"}
	case identity.CategoryCLI:
		return []string{"cli"}
	case identity.CategoryLibrary:
		return []string{"library", "lib"}
	default:
		return []string{string(c)}
	}
}

// splitBotTokens splits a --bots value on commas/whitespace into clean tokens.
func splitBotTokens(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field = strings.TrimSpace(field); field != "" {
			out = append(out, field)
		}
	}
	return out
}

func joinComma(values []string) string {
	out := ""
	for i, value := range values {
		if i > 0 {
			out += ", "
		}
		out += value
	}
	return out
}
