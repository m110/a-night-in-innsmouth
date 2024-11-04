package twine

import (
	"regexp"
	"strings"

	"github.com/m110/secrets/component"
)

// ParseStory parses the complete story text
func ParseStory(content string) (component.RawStory, error) {
	story := component.RawStory{}
	sections := strings.Split(content, "::")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		lines := strings.Split(section, "\n")
		if len(lines) == 0 {
			continue
		}

		// Parse first line to get title, tags, and metadata
		firstLine := lines[0]

		if firstLine == "StoryTitle" {
			if len(lines) > 1 {
				story.Title = strings.TrimSpace(lines[1])
			}
			continue
		}

		if firstLine == "StoryData" {
			continue
		}

		passage := parsePassage(firstLine, strings.Join(lines[1:], "\n"))
		story.Passages = append(story.Passages, passage)
	}

	return story, nil
}

// parsePassage parses a single passage section
func parsePassage(titleLine, content string) component.RawPassage {
	passage := component.RawPassage{}

	// Extract tags
	tagRegex := regexp.MustCompile(`\[(.*?)\]`)
	if matches := tagRegex.FindStringSubmatch(titleLine); len(matches) > 1 {
		tags := strings.Split(matches[1], " ")
		for _, tag := range tags {
			if tag != "" {
				passage.Tags = append(passage.Tags, tag)
			}
		}
		titleLine = tagRegex.ReplaceAllString(titleLine, "")
	}

	// Extract metadata
	metadataRegex := regexp.MustCompile(`\{(.*?)\}`)
	if matches := metadataRegex.FindStringSubmatch(titleLine); len(matches) > 1 {
		titleLine = metadataRegex.ReplaceAllString(titleLine, "")
	}

	parts := strings.Split(content, "\n--\n")
	if len(parts) > 1 {
		macros := strings.Split(parts[0], "\n")
		for _, macro := range macros {
			parts := strings.Split(macro, ":")
			if len(parts) < 2 {
				continue
			}

			macro := component.Macro{
				Type:  component.MacroType(strings.TrimSpace(parts[0])),
				Value: strings.TrimSpace(parts[1]),
			}
			passage.Macros = append(passage.Macros, macro)
		}
		content = parts[1]
	}

	passage.Title = strings.TrimSpace(titleLine)
	passage.Content = strings.TrimSpace(content)
	passage.Content, passage.Links = parseLinks(content)

	return passage
}

// parseLinks extracts links from passage content
func parseLinks(content string) (string, []component.RawLink) {
	links := []component.RawLink{}

	// Match both [[Target]] and [[Text->Target]] format links
	linkRegex := regexp.MustCompile(`(?m)^(?:> )?(\[.*?\] )*\[\[(.*?)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	content = linkRegex.ReplaceAllString(content, "")

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		link := component.RawLink{}

		tagPattern := regexp.MustCompile(`\[(.*?)\]`)
		tagMatches := tagPattern.FindAllStringSubmatch(match[1], -1)
		for _, tm := range tagMatches {
			parts := strings.SplitN(tm[1], " ", 3)

			if len(parts) < 2 {
				panic("Invalid tag format: " + tm[1])
			}

			var positive bool
			if parts[0] == "if" {
				positive = true
			} else if parts[0] != "unless" {
				panic("Invalid tag condition: " + parts[0])
			}

			cond := component.Condition{
				Positive: positive,
				Type:     component.ConditionType(parts[1]),
				Value:    parts[2],
			}

			link.Conditions = append(link.Conditions, cond)
		}

		// Check if link has display text
		parts := strings.Split(match[2], "->")
		if len(parts) > 1 {
			link.Text = strings.TrimSpace(parts[0])
			link.Target = strings.TrimSpace(parts[1])
		} else {
			link.Text = strings.TrimSpace(parts[0])
			link.Target = link.Text
		}

		links = append(links, link)
	}

	return content, links
}
