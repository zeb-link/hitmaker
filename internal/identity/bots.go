package identity

import "strings"

// BotCategory groups bots by kind (AI crawler, AI assistant, search crawler,
// link-preview fetcher, CLI client, HTTP library). Simulating a bot is entirely
// a matter of sending its real, publicly documented User-Agent string; the
// category is just how the catalog is organized.
type BotCategory string

const (
	// CategoryAICrawler is an AI training / indexing harvester (GPTBot,
	// ClaudeBot, PerplexityBot, ...).
	CategoryAICrawler BotCategory = "ai_crawler"
	// CategoryAIAssistant is an AI agent fetching a page on behalf of a live
	// user (ChatGPT-User, Claude-User, ...).
	CategoryAIAssistant BotCategory = "ai_assistant"
	// CategoryCrawler is a conventional search / SEO spider.
	CategoryCrawler BotCategory = "crawler"
	// CategoryFetcher is a link-preview unfurler (Slackbot, Twitterbot, ...).
	CategoryFetcher BotCategory = "fetcher"
	// CategoryCLI is a command-line HTTP client (curl, wget).
	CategoryCLI BotCategory = "cli"
	// CategoryLibrary is an HTTP client library (python-requests, axios, ...).
	CategoryLibrary BotCategory = "library"
)

// AllCategories lists every bot category in a stable display order.
var AllCategories = []BotCategory{
	CategoryAICrawler,
	CategoryAIAssistant,
	CategoryCrawler,
	CategoryFetcher,
	CategoryCLI,
	CategoryLibrary,
}

// CategoryDescription returns a one-line human description of a category.
func CategoryDescription(c BotCategory) string {
	switch c {
	case CategoryAICrawler:
		return "AI training/index crawlers (GPTBot, ClaudeBot, Bytespider, CCBot)"
	case CategoryAIAssistant:
		return "AI assistants fetching live for a user (ChatGPT-User, Claude-User)"
	case CategoryCrawler:
		return "Search & SEO spiders (Googlebot, Bingbot, AhrefsBot)"
	case CategoryFetcher:
		return "Link-preview unfurlers (Slackbot, Twitterbot, Discordbot)"
	case CategoryCLI:
		return "Command-line HTTP clients (curl, wget, HTTPie)"
	case CategoryLibrary:
		return "HTTP client libraries (python-requests, axios, Go-http-client)"
	default:
		return string(c)
	}
}

// Bot is one recognizable non-human agent: its canonical name, its category,
// a real public User-Agent string it sends, and a relative weight within the
// whole bot pool.
type Bot struct {
	Name     string
	Category BotCategory
	UA       string
	Weight   float64
}

// Bots is the full catalog of agents Hitmaker can impersonate. Each entry uses
// the agent's real, publicly documented User-Agent string, so analytics that
// classify traffic by bot type will recognize and categorize it.
var Bots = []Bot{
	// ── AI crawlers (ai_crawler) ──────────────────────────────
	{Name: "GPTBot", Category: CategoryAICrawler, Weight: 10, UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; GPTBot/1.1; +https://openai.com/gptbot"},
	{Name: "OAI-SearchBot", Category: CategoryAICrawler, Weight: 5, UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; OAI-SearchBot/1.0; +https://openai.com/searchbot"},
	{Name: "ClaudeBot", Category: CategoryAICrawler, Weight: 9, UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ClaudeBot/1.0; +claudebot@anthropic.com"},
	{Name: "anthropic-ai", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; anthropic-ai/1.0; +http://www.anthropic.com/bot)"},
	{Name: "Claude-Web", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; Claude-Web/1.0; +http://www.anthropic.com)"},
	{Name: "Claude-SearchBot", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; Claude-SearchBot/1.0; +https://www.anthropic.com)"},
	{Name: "PerplexityBot", Category: CategoryAICrawler, Weight: 7, UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; PerplexityBot/1.0; +https://perplexity.ai/perplexitybot"},
	{Name: "Google-Extended", Category: CategoryAICrawler, Weight: 6, UA: "Mozilla/5.0 (compatible; Google-Extended/1.0; +http://www.google.com/bot.html)"},
	{Name: "GoogleOther", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; GoogleOther)"},
	{Name: "Google-NotebookLM", Category: CategoryAICrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; Google-NotebookLM/1.0; +http://www.google.com/bot.html)"},
	{Name: "Applebot-Extended", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; Applebot-Extended/0.1; +http://www.apple.com/go/applebot)"},
	{Name: "Amazonbot", Category: CategoryAICrawler, Weight: 6, UA: "Mozilla/5.0 (compatible; Amazonbot/0.1; +https://developer.amazon.com/support/amazonbot)"},
	{Name: "Bytespider", Category: CategoryAICrawler, Weight: 9, UA: "Mozilla/5.0 (compatible; Bytespider; spider-feedback@bytedance.com)"},
	{Name: "TikTokSpider", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; TikTokSpider; +https://www.tiktok.com)"},
	{Name: "CCBot", Category: CategoryAICrawler, Weight: 6, UA: "CCBot/2.0 (https://commoncrawl.org/faq/)"},
	{Name: "meta-externalagent", Category: CategoryAICrawler, Weight: 6, UA: "meta-externalagent/1.1 (+https://developers.facebook.com/docs/sharing/webmasters/crawler)"},
	{Name: "FacebookBot", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; FacebookBot/1.0; +https://developers.facebook.com/docs/sharing/webmasters/crawler)"},
	{Name: "cohere-training-data-crawler", Category: CategoryAICrawler, Weight: 2, UA: "cohere-training-data-crawler/1.0"},
	{Name: "DeepSeekBot", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; DeepSeekBot/1.0; +https://www.deepseek.com)"},
	{Name: "DataForSeoBot", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; DataForSeoBot/1.0; +https://dataforseo.com/dataforseo-bot)"},
	{Name: "Diffbot", Category: CategoryAICrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; Diffbot/0.1; +http://www.diffbot.com)"},
	{Name: "ImagesiftBot", Category: CategoryAICrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; ImagesiftBot; +https://imagesift.com/about)"},
	{Name: "PetalBot", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; PetalBot; +https://webmaster.petalsearch.com/site/petalbot)"},
	{Name: "YouBot", Category: CategoryAICrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; YouBot (+http://about.you.com/youbot))"},
	{Name: "xAI-Bot", Category: CategoryAICrawler, Weight: 3, UA: "Mozilla/5.0 (compatible; xAI-Bot/1.0; +https://x.ai)"},
	{Name: "Timpibot", Category: CategoryAICrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; Timpibot/0.8; +http://www.timpi.io)"},
	{Name: "Webzio-Extended", Category: CategoryAICrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; Webzio-Extended/1.0)"},
	{Name: "omgili", Category: CategoryAICrawler, Weight: 1, UA: "omgili/0.5 +http://omgili.com"},
	{Name: "FirecrawlAgent", Category: CategoryAICrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; FirecrawlAgent/1.0; +https://firecrawl.dev)"},
	{Name: "AI2Bot", Category: CategoryAICrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; AI2Bot; +https://www.allenai.org/crawler)"},
	{Name: "SemrushBot-OCOB", Category: CategoryAICrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; SemrushBot-OCOB/1.0; +http://www.semrush.com/bot.html)"},
	{Name: "v0bot", Category: CategoryAICrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; v0bot/1.0; +https://v0.dev)"},
	{Name: "HuggingFace-Bot", Category: CategoryAICrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; HuggingFace-Bot/1.0)"},

	// ── AI assistants (ai_assistant) ──────────────────────────
	{Name: "ChatGPT-User", Category: CategoryAIAssistant, Weight: 9, UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot"},
	{Name: "Claude-User", Category: CategoryAIAssistant, Weight: 8, UA: "Mozilla/5.0 (compatible; Claude-User/1.0; +Claude-User@anthropic.com)"},
	{Name: "Perplexity-User", Category: CategoryAIAssistant, Weight: 6, UA: "Mozilla/5.0 (compatible; Perplexity-User/1.0; +https://perplexity.ai/perplexity-user)"},
	{Name: "MistralAI-User", Category: CategoryAIAssistant, Weight: 3, UA: "Mozilla/5.0 (compatible; MistralAI-User/1.0; +https://mistral.ai)"},
	{Name: "DuckAssistBot", Category: CategoryAIAssistant, Weight: 3, UA: "Mozilla/5.0 (compatible; DuckAssistBot/1.0; +https://duckduckgo.com/duckassistbot.html)"},
	{Name: "Gemini-Deep-Research", Category: CategoryAIAssistant, Weight: 2, UA: "Mozilla/5.0 (compatible; Gemini-Deep-Research/1.0; +http://www.google.com/bot.html)"},
	{Name: "Cohere-AI", Category: CategoryAIAssistant, Weight: 2, UA: "cohere-ai/1.0"},
	{Name: "NovaAct", Category: CategoryAIAssistant, Weight: 1, UA: "agent-novaact/1.0"},

	// ── Search & SEO crawlers (crawler) ───────────────────────
	{Name: "Googlebot", Category: CategoryCrawler, Weight: 25, UA: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"},
	{Name: "Googlebot-Mobile", Category: CategoryCrawler, Weight: 12, UA: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"},
	{Name: "bingbot", Category: CategoryCrawler, Weight: 12, UA: "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)"},
	{Name: "DuckDuckBot", Category: CategoryCrawler, Weight: 5, UA: "DuckDuckBot/1.1; (+http://duckduckgo.com/duckduckbot.html)"},
	{Name: "YandexBot", Category: CategoryCrawler, Weight: 6, UA: "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)"},
	{Name: "Baiduspider", Category: CategoryCrawler, Weight: 5, UA: "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)"},
	{Name: "Applebot", Category: CategoryCrawler, Weight: 4, UA: "Mozilla/5.0 (compatible; Applebot/0.1; +http://www.apple.com/go/applebot)"},
	{Name: "AhrefsBot", Category: CategoryCrawler, Weight: 5, UA: "Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)"},
	{Name: "SemrushBot", Category: CategoryCrawler, Weight: 4, UA: "Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)"},
	{Name: "DotBot", Category: CategoryCrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)"},
	{Name: "MJ12bot", Category: CategoryCrawler, Weight: 2, UA: "Mozilla/5.0 (compatible; MJ12bot/v1.4.8; http://mj12bot.com/)"},
	{Name: "Sogou", Category: CategoryCrawler, Weight: 2, UA: "Sogou web spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)"},
	{Name: "SeznamBot", Category: CategoryCrawler, Weight: 1, UA: "Mozilla/5.0 (compatible; SeznamBot/4.0; +http://napoveda.seznam.cz/seznambot-intro/)"},

	// ── Link-preview fetchers (fetcher) ───────────────────────
	{Name: "facebookexternalhit", Category: CategoryFetcher, Weight: 12, UA: "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"},
	{Name: "Twitterbot", Category: CategoryFetcher, Weight: 8, UA: "Twitterbot/1.0"},
	{Name: "Slackbot", Category: CategoryFetcher, Weight: 7, UA: "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)"},
	{Name: "LinkedInBot", Category: CategoryFetcher, Weight: 6, UA: "LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com)"},
	{Name: "WhatsApp", Category: CategoryFetcher, Weight: 8, UA: "WhatsApp/2.23.20.0"},
	{Name: "TelegramBot", Category: CategoryFetcher, Weight: 5, UA: "TelegramBot (like TwitterBot)"},
	{Name: "Discordbot", Category: CategoryFetcher, Weight: 6, UA: "Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)"},
	{Name: "Pinterestbot", Category: CategoryFetcher, Weight: 3, UA: "Mozilla/5.0 (compatible; Pinterestbot/1.0; +https://www.pinterest.com/bot.html)"},
	{Name: "redditbot", Category: CategoryFetcher, Weight: 3, UA: "Mozilla/5.0 (compatible; redditbot/1.0; +http://www.reddit.com/feedback)"},

	// ── Command-line clients (cli) ────────────────────────────
	{Name: "curl", Category: CategoryCLI, Weight: 8, UA: "curl/8.4.0"},
	{Name: "Wget", Category: CategoryCLI, Weight: 3, UA: "Wget/1.21.4"},
	{Name: "HTTPie", Category: CategoryCLI, Weight: 2, UA: "HTTPie/3.2.2"},
	{Name: "PowerShell", Category: CategoryCLI, Weight: 2, UA: "Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.19041.3803"},

	// ── HTTP client libraries (library) ───────────────────────
	{Name: "python-requests", Category: CategoryLibrary, Weight: 7, UA: "python-requests/2.31.0"},
	{Name: "axios", Category: CategoryLibrary, Weight: 4, UA: "axios/1.6.2"},
	{Name: "node-fetch", Category: CategoryLibrary, Weight: 3, UA: "node-fetch/3.3.2"},
	{Name: "Go-http-client", Category: CategoryLibrary, Weight: 4, UA: "Go-http-client/2.0"},
	{Name: "okhttp", Category: CategoryLibrary, Weight: 3, UA: "okhttp/4.12.0"},
	{Name: "Scrapy", Category: CategoryLibrary, Weight: 2, UA: "Scrapy/2.11.0 (+https://scrapy.org)"},
	{Name: "Java", Category: CategoryLibrary, Weight: 2, UA: "Java/17.0.9"},
	{Name: "libwww-perl", Category: CategoryLibrary, Weight: 1, UA: "libwww-perl/6.72"},
	{Name: "aiohttp", Category: CategoryLibrary, Weight: 2, UA: "Python/3.11 aiohttp/3.9.1"},
	{Name: "Bun", Category: CategoryLibrary, Weight: 1, UA: "Bun/1.0.25"},
	{Name: "Deno", Category: CategoryLibrary, Weight: 1, UA: "Deno/1.40.0"},
}

// BotFilter selects a subset of the catalog by category and/or exact name.
// An empty filter matches every bot.
type BotFilter struct {
	Categories map[BotCategory]bool
	Names      map[string]bool // lower-cased canonical names
}

// Empty reports whether the filter constrains nothing (matches all bots).
func (f BotFilter) Empty() bool {
	return len(f.Categories) == 0 && len(f.Names) == 0
}

// Match reports whether a bot passes the filter.
func (f BotFilter) Match(b Bot) bool {
	if f.Empty() {
		return true
	}
	if f.Categories[b.Category] {
		return true
	}
	return f.Names[strings.ToLower(b.Name)]
}

// BotPool returns the weighted User-Agent pool for the bots matching the filter.
// If nothing matches, it returns the full catalog so traffic never stalls.
func BotPool(filter BotFilter) []Weighted[string] {
	pool := make([]Weighted[string], 0, len(Bots))
	for _, bot := range Bots {
		if filter.Match(bot) {
			pool = append(pool, Weighted[string]{Value: bot.UA, Weight: bot.Weight})
		}
	}
	if len(pool) == 0 {
		return BotPool(BotFilter{})
	}
	return pool
}

// ParseBotSpec turns a list of tokens (category values like "ai_crawler" or "ai",
// or exact bot names like "GPTBot") into a BotFilter. It also accepts the
// convenience alias "ai" (both AI categories) and "all"/"any" (no constraint).
// Unknown tokens produce an error listing what was not recognized, so callers
// (and agents) get actionable feedback.
func ParseBotSpec(tokens []string) (BotFilter, error) {
	filter := BotFilter{Categories: map[BotCategory]bool{}, Names: map[string]bool{}}
	var unknown []string
	names := botNameIndex()
	for _, raw := range tokens {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}
		key := strings.ToLower(token)
		switch key {
		case "all", "any", "*":
			// No constraint; leave categories/names empty.
			continue
		case "ai":
			filter.Categories[CategoryAICrawler] = true
			filter.Categories[CategoryAIAssistant] = true
			continue
		}
		if cat, ok := categoryAlias(key); ok {
			filter.Categories[cat] = true
			continue
		}
		if canonical, ok := names[key]; ok {
			filter.Names[strings.ToLower(canonical)] = true
			continue
		}
		unknown = append(unknown, token)
	}
	if len(unknown) > 0 {
		return BotFilter{}, &UnknownBotError{Tokens: unknown}
	}
	return filter, nil
}

// UnknownBotError reports bot-spec tokens that matched no category or name.
type UnknownBotError struct {
	Tokens []string
}

func (e *UnknownBotError) Error() string {
	return "unknown bot category or name: " + strings.Join(e.Tokens, ", ") +
		" (run `hitmaker bots` to list valid categories and names)"
}

func categoryAlias(key string) (BotCategory, bool) {
	switch key {
	case "ai_crawler", "ai-crawler", "aicrawler":
		return CategoryAICrawler, true
	case "ai_assistant", "ai-assistant", "aiassistant", "assistant":
		return CategoryAIAssistant, true
	case "crawler", "crawlers", "search":
		return CategoryCrawler, true
	case "fetcher", "fetchers", "social", "preview":
		return CategoryFetcher, true
	case "cli", "clis":
		return CategoryCLI, true
	case "library", "libraries", "lib":
		return CategoryLibrary, true
	default:
		return "", false
	}
}

func botNameIndex() map[string]string {
	index := make(map[string]string, len(Bots))
	for _, bot := range Bots {
		index[strings.ToLower(bot.Name)] = bot.Name
	}
	return index
}

// BotsByCategory groups the catalog by category in AllCategories order.
func BotsByCategory() map[BotCategory][]Bot {
	grouped := map[BotCategory][]Bot{}
	for _, bot := range Bots {
		grouped[bot.Category] = append(grouped[bot.Category], bot)
	}
	return grouped
}
