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

	passage.Title = strings.TrimSpace(titleLine)
	passage.Content = strings.TrimSpace(content)
	passage.Content, passage.Links = parseLinks(content)

	return passage
}

// parseLinks extracts links from passage content
func parseLinks(content string) (string, []component.RawLink) {
	links := []component.RawLink{}

	// Match both [[Target]] and [[Text->Target]] format links
	linkRegex := regexp.MustCompile(`(?m)^(?:> )?\[\[(.*?)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	content = linkRegex.ReplaceAllString(content, "")

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		link := component.RawLink{}

		// Check if link has display text
		parts := strings.Split(match[1], "->")
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
