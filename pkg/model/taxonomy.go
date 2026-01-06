package model

// Channel represents the primary classification of a beat
type Channel int

const (
	ChannelUnknown Channel = iota
	ChannelCoaching     // Insights from coaching/mentoring
	ChannelResearch     // Deliberate investigation
	ChannelDiscovery    // Serendipitous finding
	ChannelDevelopment  // Building/coding insight
	ChannelReflection   // Personal synthesis
	ChannelReference    // Saved for later use
	ChannelMilestone    // Achievement/completion
)

func (c Channel) String() string {
	switch c {
	case ChannelCoaching:
		return "Coaching"
	case ChannelResearch:
		return "Research"
	case ChannelDiscovery:
		return "Discovery"
	case ChannelDevelopment:
		return "Development"
	case ChannelReflection:
		return "Reflection"
	case ChannelReference:
		return "Reference"
	case ChannelMilestone:
		return "Milestone"
	default:
		return "Unknown"
	}
}

// AllChannels returns all valid channels
func AllChannels() []Channel {
	return []Channel{
		ChannelCoaching,
		ChannelResearch,
		ChannelDiscovery,
		ChannelDevelopment,
		ChannelReflection,
		ChannelReference,
		ChannelMilestone,
	}
}

// Source represents the origin type of a beat
type Source int

const (
	SourceUnknown Source = iota
	SourceConversation // Human dialogue
	SourceWeb          // Browser discovery
	SourceTwitter      // X/Twitter
	SourceGitHub       // Code/issues/discussions
	SourceBook         // Reading
	SourceSession      // Agent/droid session
	SourceInternal     // Self-generated
)

func (s Source) String() string {
	switch s {
	case SourceConversation:
		return "Conversation"
	case SourceWeb:
		return "Web"
	case SourceTwitter:
		return "Twitter"
	case SourceGitHub:
		return "GitHub"
	case SourceBook:
		return "Book"
	case SourceSession:
		return "Session"
	case SourceInternal:
		return "Internal"
	default:
		return "Unknown"
	}
}

// AllSources returns all valid sources
func AllSources() []Source {
	return []Source{
		SourceConversation,
		SourceWeb,
		SourceTwitter,
		SourceGitHub,
		SourceBook,
		SourceSession,
		SourceInternal,
	}
}

// Taxonomy represents the classification of a beat
type Taxonomy struct {
	Channel    Channel `json:"channel"`
	Source     Source  `json:"source"`
	Confidence float64 `json:"confidence"`
}
