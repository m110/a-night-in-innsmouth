package twine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m110/secrets/component"
)

func TestParsePassage(t *testing.T) {
	passage := `addItem: key
takeMoney: 100
--
First line.

Second line.
And third line.

[if hasItem key]
Only if has key.
[continue]

This is always visible.

[if hasItem key]
Only if has key.
[else]
Only if no key.
[continue]

[if hasItem key && !hasMoney 100]
Only if has key and not 100 money.

[if !hasMoney 100]
> [[Poor man]]
[else]
> [[Rich man]]
[continue]

[if !hasMoney 200]
Only if not 200 money.
[continue]

[if hasItem key]
> {back} [[Leave Key->No Key]]
[continue]

> [[Exit]]
`

	parsed := parsePassage("This is title [tag1 tag2]", passage)

	expectedPassage := component.RawPassage{
		Title: "This is title",
		Tags:  []string{"tag1", "tag2"},
		Segments: []component.Segment{
			{
				Text: "First line.\n\nSecond line.\nAnd third line.",
			},
			{
				Text: "Only if has key.",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasItem,
						Value:    "key",
					},
				},
			},
			{
				Text: "This is always visible.",
			},
			{
				Text: "Only if has key.",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasItem,
						Value:    "key",
					},
				},
			},
			{
				Text: "Only if no key.",
				Conditions: []component.Condition{
					{
						Positive: false,
						Type:     component.ConditionTypeHasItem,
						Value:    "key",
					},
				},
			},
			{
				Text: "Only if has key and not 100 money.",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasItem,
						Value:    "key",
					},
					{
						Positive: false,
						Type:     component.ConditionTypeHasMoney,
						Value:    "100",
					},
				},
			},
			{
				Text: "Only if not 200 money.",
				Conditions: []component.Condition{
					{
						Positive: false,
						Type:     component.ConditionTypeHasMoney,
						Value:    "200",
					},
				},
			},
		},
		Macros: []component.Macro{
			{
				Type:  component.MacroTypeAddItem,
				Value: "key",
			},
			{
				Type:  component.MacroTypeTakeMoney,
				Value: "100",
			},
		},
		Links: []component.RawLink{
			{
				Text:   "Poor man",
				Target: "Poor man",
				Conditions: []component.Condition{
					{
						Positive: false,
						Type:     component.ConditionTypeHasMoney,
						Value:    "100",
					},
				},
			},
			{
				Text:   "Rich man",
				Target: "Rich man",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasMoney,
						Value:    "100",
					},
				},
			},
			{
				Text:   "Leave Key",
				Target: "No Key",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasItem,
						Value:    "key",
					},
				},
				Tags: []string{"back"},
			},
			{
				Text:   "Exit",
				Target: "Exit",
			},
		},
	}

	assert.Equal(t, expectedPassage, parsed)
}

func TestParsePassage_Story(t *testing.T) {
	passage := `[if fact day2]
It's early morning. The town is still quiet

[else]
The sun is getting low.

[continue]

> [[Go to the train station->Train Station]]
> [[Talk with the Shopkeeper->Newburyport Shopkeeper]]

[if hasItem Arkham Daily Newspaper]

> [[Read the newspaper]]

[continue]

[if fact day2]

> [[Go to Bus Station->Bus Station]]

[else]

> [[Visit your room at YMCA->YMCA]]

[continue]
`

	parsed := parsePassage("This is title [tag1 tag2]", passage)

	expectedPassage := component.RawPassage{
		Title: "This is title",
		Tags:  []string{"tag1", "tag2"},
		Segments: []component.Segment{
			{
				Text: "It's early morning. The town is still quiet",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeFact,
						Value:    "day2",
					},
				},
			},
			{
				Text: "The sun is getting low.",
				Conditions: []component.Condition{
					{
						Positive: false,
						Type:     component.ConditionTypeFact,
						Value:    "day2",
					},
				},
			},
		},
		Links: []component.RawLink{
			{
				Text:   "Go to the train station",
				Target: "Train Station",
			},
			{
				Text:   "Talk with the Shopkeeper",
				Target: "Newburyport Shopkeeper",
			},
			{
				Text:   "Read the newspaper",
				Target: "Read the newspaper",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasItem,
						Value:    "Arkham Daily Newspaper",
					},
				},
			},
			{
				Text:   "Go to Bus Station",
				Target: "Bus Station",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeFact,
						Value:    "day2",
					},
				},
			},
			{
				Text:   "Visit your room at YMCA",
				Target: "YMCA",
				Conditions: []component.Condition{
					{
						Positive: false,
						Type:     component.ConditionTypeFact,
						Value:    "day2",
					},
				},
			},
		},
	}

	assert.Equal(t, expectedPassage, parsed)
}

func TestParsePassage_Story2(t *testing.T) {
	passage := `You arrive at the train station.

[unless hasItem Train Ticket to Arkham]
> [[Check tickets to Arkham]]
[continue]

[if hasItem Train Ticket to Arkham && fact day2]
>  [[Board the train to Arkham]] 
[continue]

> [[Talk to the Agent->Agent]]
> [[Exit->Newburyport]] 
`

	parsed := parsePassage("Train Station", passage)

	expectedPassage := component.RawPassage{
		Title: "Train Station",
		Segments: []component.Segment{
			{
				Text: "You arrive at the train station.",
			},
		},
		Links: []component.RawLink{
			{
				Text:   "Check tickets to Arkham",
				Target: "Check tickets to Arkham",
				Conditions: []component.Condition{
					{
						Positive: false,
						Type:     component.ConditionTypeHasItem,
						Value:    "Train Ticket to Arkham",
					},
				},
			},
			{
				Text:   "Board the train to Arkham",
				Target: "Board the train to Arkham",
				Conditions: []component.Condition{
					{
						Positive: true,
						Type:     component.ConditionTypeHasItem,
						Value:    "Train Ticket to Arkham",
					},
					{
						Positive: true,
						Type:     component.ConditionTypeFact,
						Value:    "day2",
					},
				},
			},
			{
				Text:   "Talk to the Agent",
				Target: "Agent",
			},
			{
				Text:   "Exit",
				Target: "Newburyport",
			},
		},
	}

	assert.Equal(t, expectedPassage, parsed)
}
