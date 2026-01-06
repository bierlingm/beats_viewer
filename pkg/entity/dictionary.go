package entity

import "github.com/bierlingm/beats_viewer/pkg/model"

// Known people dictionary - common tech/startup figures
var KnownPeople = []string{
	"DHH", "Claude", "David", "Simon",
	"Paul Graham", "Patrick Collison", "Sam Altman",
}

// Known tools dictionary
var KnownTools = []string{
	"Supabase", "Ollama", "GitHub", "Cloudflare", "Vercel",
	"beads", "beats", "bv", "btv", "bd", "Factory", "Droid",
	"Claude", "ChatGPT", "GPT", "Cursor", "VSCode",
	"React", "Next.js", "TypeScript", "Go", "Python", "Rust",
	"Docker", "Kubernetes", "AWS", "GCP", "PostgreSQL", "Redis",
	"Notion", "Linear", "Slack", "Discord", "Figma",
	"WezTerm", "tmux", "nvim", "vim", "git",
}

// Known concepts dictionary
var KnownConcepts = []string{
	"commitment", "identity", "narrative substrate", "psychoid buffer",
	"agent", "agentic", "workflow", "automation", "flywheel",
	"synthesis", "pattern", "insight", "discovery",
}

// Known projects (customize for your context)
var KnownProjects = []string{
	"runcible", "modern-minuteman", "modern minuteman",
}

// Known organizations
var KnownOrganizations = []string{
	"Factory", "Anthropic", "OpenAI", "Google", "Meta", "Microsoft",
	"Stripe", "Vercel", "Cloudflare", "37signals", "Basecamp",
}

// EntityDictionaries groups all dictionaries by type
var EntityDictionaries = map[model.EntityType][]string{
	model.EntityPerson:       KnownPeople,
	model.EntityTool:         KnownTools,
	model.EntityConcept:      KnownConcepts,
	model.EntityProject:      KnownProjects,
	model.EntityOrganization: KnownOrganizations,
}
