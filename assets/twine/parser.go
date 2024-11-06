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
				Type:  parseMacroType(strings.TrimSpace(parts[0])),
				Value: strings.TrimSpace(parts[1]),
			}
			passage.Macros = append(passage.Macros, macro)
		}
		content = parts[1]
	}

	passage.Title = strings.TrimSpace(titleLine)
	content = strings.TrimSpace(content)

	var segments []component.Segment
	currentSegment := component.Segment{}
	var currentConditions []component.Condition
	conditionStarted := false

	for _, segment := range strings.Split(content, "\n") {
		if segment == "[else]" {
			if !conditionStarted {
				panic("Invalid [else] tag")
			}

			if len(currentConditions) == 0 {
				panic("Missing else condition")
			}

			if currentSegment.Text != "" {
				segments = append(segments, currentSegment)
			}
			currentSegment = component.Segment{}
			for _, cond := range currentConditions {
				cond.Positive = !cond.Positive
				currentSegment.Conditions = append(currentSegment.Conditions, cond)
			}
			currentConditions = nil
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
			currentConditions = nil
			continue
		}

		if strings.HasPrefix(segment, "[if") || strings.HasPrefix(segment, "[unless") {
			if conditionStarted {
				if currentSegment.Text != "" {
					segments = append(segments, currentSegment)
				}
				currentSegment = component.Segment{}
				conditionStarted = false
				currentConditions = nil
			}

			if currentSegment.Text != "" {
				segments = append(segments, currentSegment)
			}
			currentSegment = component.Segment{}
			conditionStarted = true

			currentConditions = parseConditions(segment)

			currentSegment.Conditions = append(currentSegment.Conditions, currentConditions...)
			continue
		}

		currentSegment.Text += segment + "\n"
	}

	if currentSegment.Text != "" {
		segments = append(segments, currentSegment)
	}

	passage.Segments, passage.Links = parseLinks(segments)

	return passage
}

func parseConditions(str string) []component.Condition {
	var conditions []component.Condition

	str = strings.Trim(str, "[]")
	firstWord := strings.Split(str, " ")[0]

	var positive bool
	if firstWord == "if" {
		positive = true
	} else if firstWord != "unless" {
		panic("Invalid tag condition: " + firstWord + " " + str)
	}

	str = strings.Trim(str[len(firstWord):], " ")

	for _, cond := range strings.Split(str, "&&") {
		parts := strings.SplitN(strings.TrimSpace(cond), " ", 2)

		condType := parts[0]

		condPositive := positive
		if strings.HasPrefix(condType, "!") {
			condPositive = !condPositive
			condType = condType[1:]
		}

		c := component.Condition{
			Positive: condPositive,
			Type:     parseConditionType(condType),
			Value:    parts[1],
		}
		conditions = append(conditions, c)
	}
	return conditions
}

// Match both [[Target]] and [[Text->Target]] format links
var linkRegex = regexp.MustCompile(`(?m)^(?:>\s+)?(\{.*?\} )?(\[.*?\] )*\[\[(.*?)\]\]`)

func parseLinks(segments []component.Segment) ([]component.Segment, []component.RawLink) {
	finalSegments := []component.Segment{}
	links := []component.RawLink{}

	for _, segment := range segments {
		matches := linkRegex.FindAllStringSubmatch(segment.Text, -1)
		segment.Text = strings.TrimSpace(linkRegex.ReplaceAllString(segment.Text, ""))

		for _, match := range matches {
			if len(match) < 4 {
				continue
			}

			link := component.RawLink{
				Conditions: segment.Conditions,
			}

			tagPattern := regexp.MustCompile(`\{(.*?)\}`)
			tagMatches := tagPattern.FindAllStringSubmatch(match[1], -1)
			for _, tm := range tagMatches {
				link.Tags = strings.Fields(tm[1])
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

		if segment.Text != "" {
			finalSegments = append(finalSegments, segment)
		}
	}

	return finalSegments, links
}

func parseMacroType(str string) component.MacroType {
	switch str {
	case "addItem":
		return component.MacroTypeAddItem
	case "takeItem":
		return component.MacroTypeTakeItem
	case "addMoney":
		return component.MacroTypeAddMoney
	case "takeMoney":
		return component.MacroTypeTakeMoney
	case "addFact":
		return component.MacroTypeAddFact
	default:
		panic("Invalid macro type: " + str)
	}
}

func parseConditionType(str string) component.ConditionType {
	switch str {
	case "hasItem":
		return component.ConditionTypeHasItem
	case "hasMoney":
		return component.ConditionTypeHasMoney
	case "fact":
		return component.ConditionTypeFact
	default:
		panic("Invalid condition type: " + str)
	}
}
