package taxonomy

import "beats_viewer/pkg/model"

// ChannelPatterns maps channels to detection patterns
var ChannelPatterns = map[model.Channel][]string{
	model.ChannelCoaching:    {"coaching", "nick", "mentor", "insight from", "advice", "guidance", "feedback"},
	model.ChannelResearch:    {"research", "study", "paper", "investigation", "analysis", "exploring", "learned"},
	model.ChannelDiscovery:   {"discovery", "found", "discovered", "stumbled", "noticed", "came across", "interesting"},
	model.ChannelDevelopment: {"development", "built", "implemented", "code", "programming", "shipped", "deployed", "refactor"},
	model.ChannelReflection:  {"reflection", "thinking", "realized", "synthesis", "pondering", "contemplating", "insight"},
	model.ChannelReference:   {"reference", "bookmark", "save", "purchase", "reminder", "note to self", "later"},
	model.ChannelMilestone:   {"milestone", "complete", "shipped", "published", "finished", "achieved", "launched"},
}

// SourcePatterns maps sources to detection patterns
var SourcePatterns = map[model.Source][]string{
	model.SourceTwitter:      {"twitter", "x discovery", "tweet", "@", "x.com"},
	model.SourceGitHub:       {"github", "repo", "issue", "pr", "pull request", "commit"},
	model.SourceWeb:          {"web", "article", "blog", "site", "http", "url", "link"},
	model.SourceConversation: {"coaching", "call", "chat", "conversation", "meeting", "discussion", "talked"},
	model.SourceBook:         {"book", "reading", "chapter", "author", "page"},
	model.SourceSession:      {"session", "droid", "factory", "agent", "claude"},
	model.SourceInternal:     {"thinking", "reflection", "realized", "insight", "idea"},
}

// MetaChannelMap maps meta["channel"] values to Source
var MetaChannelMap = map[string]model.Source{
	"twitter":     model.SourceTwitter,
	"x":           model.SourceTwitter,
	"github":      model.SourceGitHub,
	"web":         model.SourceWeb,
	"browser":     model.SourceWeb,
	"book":        model.SourceBook,
	"reading":     model.SourceBook,
	"session":     model.SourceSession,
	"agent":       model.SourceSession,
	"droid":       model.SourceSession,
	"coaching":    model.SourceConversation,
	"call":        model.SourceConversation,
	"internal":    model.SourceInternal,
	"reflection":  model.SourceInternal,
}
