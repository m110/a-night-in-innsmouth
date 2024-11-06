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
	content = strings.TrimSpace(content)
	content, passage.Links = parseLinks(content)

	var segments []component.Segment
	currentSegment := component.Segment{}
	var currentCondition *component.Condition
	conditionStarted := false

	for _, segment := range strings.Split(content, "\n") {
		if segment == "[else]" {
			if !conditionStarted {
				panic("Invalid [else] tag")
			}

			if currentCondition == nil {
				panic("Missing else condition")
			}

			if currentSegment.Text != "" {
				segments = append(segments, currentSegment)
			}
			currentSegment = component.Segment{}
			currentCondition.Positive = !currentCondition.Positive
			currentSegment.Conditions = append(currentSegment.Conditions, *currentCondition)
			currentCondition = nil
			continue
		}

		if segment == "[continue]" {
			if !conditionStarted {
				panic("Invalid [continue] tag")
			}

			if currentSegment.Text != "" {
				segments = append(segments, currentSegment)
			}
			currentSegment = component.Segment{}
			conditionStarted = false
			currentCondition = nil
			continue
		}

		if strings.HasPrefix(segment, "[") {
			if conditionStarted {
				panic("Invalid tag inside condition: " + segment)
			}

			if currentSegment.Text != "" {
				segments = append(segments, currentSegment)
			}
			currentSegment = component.Segment{}
			conditionStarted = true

			condition := parseCondition(segment)
			currentCondition = &condition

			currentSegment.Conditions = append(currentSegment.Conditions, condition)
			continue
		}

		currentSegment.Text += segment + "\n"
	}

	if currentSegment.Text != "" {
		segments = append(segments, currentSegment)
	}

	passage.Segments = segments

	return passage
}

func parseCondition(str string) component.Condition {
	parts := strings.SplitN(strings.Trim(str, "[]"), " ", 3)

	var positive bool
	if parts[0] == "if" {
		positive = true
	} else if parts[0] != "unless" {
		panic("Invalid tag condition: " + parts[0])
	}

	return component.Condition{
		Positive: positive,
		Type:     component.ConditionType(parts[1]),
		Value:    parts[2],
	}
}

// parseLinks extracts links from passage content
func parseLinks(content string) (string, []component.RawLink) {
	links := []component.RawLink{}

	// Match both [[Target]] and [[Text->Target]] format links
	linkRegex := regexp.MustCompile(`(?m)^(?:> )?(\{.*?\} )?(\[.*?\] )*\[\[(.*?)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	content = linkRegex.ReplaceAllString(content, "")

	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		link := component.RawLink{}

		tagPattern := regexp.MustCompile(`\{(.*?)\}`)
		tagMatches := tagPattern.FindAllStringSubmatch(match[1], -1)
		for _, tm := range tagMatches {
			link.Tags = strings.Fields(tm[1])
		}

		conditionPattern := regexp.MustCompile(`\[(.*?)\]`)
		conditionMatches := conditionPattern.FindAllStringSubmatch(match[2], -1)
		for _, cm := range conditionMatches {
			parts := strings.SplitN(cm[1], " ", 3)

			if len(parts) < 2 {
				panic("Invalid tag format: " + cm[1])
			}

			var positive bool
			if parts[0] == "if" {
				positive = true
			} else if parts[0] != "unless" {
				panic("Invalid condition: " + parts[0])
			}

			cond := component.Condition{
				Positive: positive,
				Type:     component.ConditionType(parts[1]),
				Value:    parts[2],
			}

			link.Conditions = append(link.Conditions, cond)
		}

		// Check if link has display text
		parts := strings.Split(match[3], "->")
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
