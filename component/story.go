package component

type RawStory struct {
	Title    string
	Passages []RawPassage
}

type RawPassage struct {
	Title   string
	Tags    []string
	Content string
	Macros  []Macro
	Links   []RawLink
}

type RawLink struct {
	Text       string
	Target     string
	Conditions []Condition
}

type Story struct {
	Title    string
	Passages map[string]*Passage

	Items map[string]int
	Facts map[string]struct{}
}

func NewStory(rawStory RawStory) *Story {
	story := &Story{
		Title: rawStory.Title,
	}

	passagesMap := map[string]*Passage{}
	for _, p := range rawStory.Passages {
		var links []Link
		for _, l := range p.Links {
			links = append(links, Link{
				Text:       l.Text,
				Conditions: l.Conditions,
			})
		}

		var isOneTime bool
		for _, tag := range p.Tags {
			if tag == "once" {
				isOneTime = true
			}
		}

		passagesMap[p.Title] = &Passage{
			story:     story,
			Title:     p.Title,
			Content:   p.Content,
			Macros:    p.Macros,
			AllLinks:  links,
			IsOneTime: isOneTime,
		}
	}

	for _, p := range rawStory.Passages {
		for i, l := range p.Links {
			passagesMap[p.Title].AllLinks[i].Target = passagesMap[l.Target]
		}
	}

	story.Passages = passagesMap
	story.Items = map[string]int{}
	story.Facts = map[string]struct{}{}

	return story
}

func (s *Story) PassageByTitle(title string) *Passage {
	p, ok := s.Passages[title]
	if !ok {
		panic("Passage not found: " + title)
	}

	return p
}

func (s *Story) AddItem(item string) {
	s.Items[item]++
}

func (s *Story) AddFact(fact string) {
	s.Facts[fact] = struct{}{}
}

func (s *Story) TestCondition(c Condition) bool {
	switch c.Type {
	case ConditionTypeHasItem:
		_, ok := s.Items[c.Value]
		return ok == c.Positive
	case ConditionTypeFact:
		_, ok := s.Facts[c.Value]
		return ok == c.Positive
	}

	return false
}

type Passage struct {
	story *Story

	Title    string
	Content  string
	Macros   []Macro
	AllLinks []Link

	IsOneTime bool
	Visited   bool
}

func (p *Passage) Visit() {
	p.Visited = true

	for _, m := range p.Macros {
		switch m.Type {
		case MacroTypeAddItem:
			p.story.AddItem(m.Value)
		case MacroTypeAddFact:
			p.story.AddFact(m.Value)
		}
	}
}

func (p *Passage) Links() []Link {
	var links []Link
	for _, l := range p.AllLinks {
		if l.Target.IsOneTime && l.Target.Visited {
			continue
		}

		var skip bool
		for _, c := range l.Conditions {
			if !p.story.TestCondition(c) {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		links = append(links, l)
	}

	return links
}

type Link struct {
	Text       string
	Target     *Passage
	Conditions []Condition
}

type MacroType string

const (
	MacroTypeAddItem MacroType = "addItem"
	MacroTypeAddFact MacroType = "addFact"
)

type Macro struct {
	Type  MacroType
	Value string
}

type ConditionType string

const (
	ConditionTypeHasItem ConditionType = "hasItem"
	ConditionTypeFact    ConditionType = "fact"
)

type Condition struct {
	Positive bool
	Type     ConditionType
	Value    string
}
