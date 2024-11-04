package component

type RawStory struct {
	Title    string
	Passages []RawPassage
}

type RawPassage struct {
	Title   string
	Tags    []string
	Content string
	Links   []RawLink
}

type RawLink struct {
	Text   string
	Target string
}

type Story struct {
	Title    string
	Passages map[string]*Passage
}

func NewStory(rawStory RawStory) *Story {
	passagesMap := map[string]*Passage{}
	for _, p := range rawStory.Passages {
		var links []Link
		for _, l := range p.Links {
			links = append(links, Link{
				Text: l.Text,
			})
		}

		var isOneTime bool
		for _, tag := range p.Tags {
			if tag == "once" {
				isOneTime = true
			}
		}

		passagesMap[p.Title] = &Passage{
			Title:     p.Title,
			Content:   p.Content,
			AllLinks:  links,
			IsOneTime: isOneTime,
		}
	}

	for _, p := range rawStory.Passages {
		for i, l := range p.Links {
			passagesMap[p.Title].AllLinks[i].Target = passagesMap[l.Target]
		}
	}

	return &Story{
		Title:    rawStory.Title,
		Passages: passagesMap,
	}
}

func (s Story) PassageByTitle(title string) *Passage {
	p, ok := s.Passages[title]
	if !ok {
		panic("Passage not found: " + title)
	}

	return p
}

type Passage struct {
	Title    string
	Content  string
	AllLinks []Link

	IsOneTime bool
	Visited   bool
}

func (p *Passage) Links() []Link {
	var links []Link
	for _, l := range p.AllLinks {
		if l.Target.IsOneTime && l.Target.Visited {
			continue
		}
		links = append(links, l)
	}

	return links
}

type Link struct {
	Text   string
	Target *Passage
}
