package twine

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/m110/secrets/domain"
)

// ParseStory parses the complete story text
func ParseStory(content string) (domain.RawStory, error) {
	story := domain.RawStory{}
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

type expression struct {
	command   string
	arguments []string
}

func (e expression) String() string {
	exp := []string{e.command}
	exp = append(exp, e.arguments...)
	return strings.Join(exp, " ")
}

func parsePassage(titleLine, content string) domain.RawPassage {
	passage := domain.RawPassage{}

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

	if strings.HasSuffix(content, "\n--") {
		// Special case: passage with no content, just macros
		content += "\n"
	}

	parts := strings.Split(content, "\n--\n")
	if len(parts) > 1 {
		macros := strings.Split(parts[0], "\n")
		for _, macro := range macros {
			parts := strings.Split(macro, ":")
			if len(parts) < 2 {
				panic("Invalid macro: " + macro)
			}

			macroType := strings.TrimSpace(parts[0])
			macroValue := strings.TrimSpace(parts[1])

			if macroType == "if" || macroType == "unless" {
				passage.Conditions = parseConditions(macroType + " " + macroValue)
				continue
			}

			macro := domain.Macro{
				Type:  parseMacroType(macroType),
				Value: macroValue,
			}
			passage.Macros = append(passage.Macros, macro)
		}
		content = parts[1]
	}

	passage.Title = strings.TrimSpace(titleLine)
	content = strings.TrimSpace(content)

	var paragraphs []domain.Paragraph
	currentParagraph := domain.Paragraph{}
	var currentConditions []domain.Condition
	paragraphStarted := false

	startParagraphIfNotStarted := func() {
		if !paragraphStarted {
			if currentParagraph.Text != "" {
				paragraphs = append(paragraphs, currentParagraph)
			}
			currentParagraph = domain.Paragraph{}
			paragraphStarted = true
		}
	}

	for _, line := range strings.Split(content, "\n") {
		expressions := parseExpressions(line)

		if len(expressions) == 0 {
			currentParagraph.Text += line + "\n"

			// Double newline indicates end of paragraph
			if !paragraphStarted &&
				strings.HasSuffix(currentParagraph.Text, "\n\n") &&
				strings.TrimSpace(currentParagraph.Text) != "" {
				paragraphs = append(paragraphs, currentParagraph)
				currentParagraph = domain.Paragraph{}
			}

			continue
		}

		for _, exp := range expressions {
			switch exp.command {
			case "center":
				startParagraphIfNotStarted()
				currentParagraph.Align = domain.ParagraphAlignCenter
			case "h1":
				startParagraphIfNotStarted()
				currentParagraph.Type = domain.ParagraphTypeHeader
			case "hint":
				startParagraphIfNotStarted()
				currentParagraph.Type = domain.ParagraphTypeHint
			case "fear":
				startParagraphIfNotStarted()
				currentParagraph.Type = domain.ParagraphTypeFear
			case "after":
				startParagraphIfNotStarted()

				if len(exp.arguments) != 1 {
					panic("expected one argument to after")
				}

				duration, err := time.ParseDuration(exp.arguments[0])
				if err != nil {
					panic("error parsing duration: " + err.Error())
				}

				currentParagraph.Delay = duration
			case "else":
				if !paragraphStarted {
					panic("Invalid [else] tag")
				}

				if len(currentConditions) == 0 {
					panic("Missing else condition")
				}

				if currentParagraph.Text != "" {
					paragraphs = append(paragraphs, currentParagraph)
				}
				currentParagraph = domain.Paragraph{}
				for _, cond := range currentConditions {
					cond.Positive = !cond.Positive
					currentParagraph.Conditions = append(currentParagraph.Conditions, cond)
				}
				currentConditions = nil
			case "continue":
				if !paragraphStarted {
					panic("Invalid [continue] tag")
				}

				if currentParagraph.Text != "" {
					paragraphs = append(paragraphs, currentParagraph)
				}
				currentParagraph = domain.Paragraph{}
				paragraphStarted = false
				currentConditions = nil
			case "if", "unless":
				if paragraphStarted {
					if currentParagraph.Text != "" {
						paragraphs = append(paragraphs, currentParagraph)
					}
					currentParagraph = domain.Paragraph{}
					paragraphStarted = false
					currentConditions = nil
				}

				if currentParagraph.Text != "" {
					paragraphs = append(paragraphs, currentParagraph)
				}
				currentParagraph = domain.Paragraph{}
				paragraphStarted = true

				currentConditions = parseConditions(exp.String())

				currentParagraph.Conditions = append(currentParagraph.Conditions, currentConditions...)
			default:
				panic("unknown command: " + exp.command)
			}
		}
	}

	currentParagraph.Text = strings.TrimSpace(currentParagraph.Text)
	if currentParagraph.Text != "" {
		paragraphs = append(paragraphs, currentParagraph)
	}

	passage.Paragraphs, passage.Links = parseLinks(paragraphs)

	return passage
}

func parseExpressions(line string) []expression {
	var expressions []expression

	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		expressionLine := strings.Trim(line, "[]")
		for _, e := range strings.Split(expressionLine, ",") {
			e = strings.TrimSpace(e)
			expParts := strings.SplitN(e, " ", 2)

			exp := expression{
				command: expParts[0],
			}

			if len(expParts) > 1 {
				exp.arguments = expParts[1:]
			}

			expressions = append(expressions, exp)
		}
	}

	return expressions
}

func parseConditions(str string) []domain.Condition {
	var conditions []domain.Condition

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

		c := domain.Condition{
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

func parseLinks(paragraphs []domain.Paragraph) ([]domain.Paragraph, []domain.RawLink) {
	finalParagraphs := []domain.Paragraph{}
	links := []domain.RawLink{}

	for _, paragraph := range paragraphs {
		matches := linkRegex.FindAllStringSubmatch(paragraph.Text, -1)
		paragraph.Text = strings.TrimSpace(linkRegex.ReplaceAllString(paragraph.Text, ""))

		for _, match := range matches {
			if len(match) < 4 {
				continue
			}

			link := domain.RawLink{
				Conditions: paragraph.Conditions,
			}

			tagPattern := regexp.MustCompile(`\{(.*?)\}`)
			tagMatches := tagPattern.FindAllStringSubmatch(match[1], -1)
			for _, tm := range tagMatches {
				if strings.Contains(tm[1], ":") {
					parts := strings.Split(tm[1], ":")
					if strings.TrimSpace(parts[0]) == "level" {
						level := strings.Split(parts[1], ",")
						var entrypoint *int
						if len(level) == 2 {
							e, err := strconv.Atoi(strings.TrimSpace(level[1]))
							if err != nil {
								panic(err)
							}
							entrypoint = &e
						} else if len(level) > 2 {
							panic("invalid level: " + parts[1])
						}

						link.Level = &domain.TargetLevel{
							Name:       strings.TrimSpace(level[0]),
							Entrypoint: entrypoint,
						}
					}
				} else {
					link.Tags = strings.Fields(tm[1])
				}
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

		if paragraph.Text != "" {
			finalParagraphs = append(finalParagraphs, paragraph)
		}
	}

	return finalParagraphs, links
}

func parseMacroType(str string) domain.MacroType {
	switch str {
	case "addItem":
		return domain.MacroTypeAddItem
	case "takeItem":
		return domain.MacroTypeTakeItem
	case "addMoney":
		return domain.MacroTypeAddMoney
	case "takeMoney":
		return domain.MacroTypeTakeMoney
	case "setFact":
		return domain.MacroTypeSetFact
	case "playMusic":
		return domain.MacroTypePlayMusic
	case "changeCharacterSpeed":
		return domain.MacroTypeChangeCharacterSpeed
	default:
		panic("Invalid macro type: " + str)
	}
}

func parseConditionType(str string) domain.ConditionType {
	switch str {
	case "hasItem":
		return domain.ConditionTypeHasItem
	case "hasMoney":
		return domain.ConditionTypeHasMoney
	case "fact":
		return domain.ConditionTypeFact
	default:
		panic("Invalid condition type: " + str)
	}
}
