package skill

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type PromptSection struct {
	Trigger  string
	Content  string
	Required bool
}

type ProgressivePrompt struct {
	Sections   []PromptSection
	BasePrompt string
}

func LoadProgressivePrompt(promptPath string) (*ProgressivePrompt, error) {
	file, err := os.Open(promptPath)
	if err != nil {
		return nil, fmt.Errorf("open prompt file: %w", err)
	}
	defer file.Close()

	var basePrompt strings.Builder
	var sections []PromptSection
	var currentSection *PromptSection

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "# @section ") {
			if currentSection != nil {
				sections = append(sections, *currentSection)
			}
			trigger := strings.TrimSpace(strings.TrimPrefix(line, "# @section "))
			currentSection = &PromptSection{
				Trigger:  trigger,
				Required: strings.HasPrefix(trigger, "*"),
			}
			if currentSection.Required {
				currentSection.Trigger = strings.TrimPrefix(trigger, "*")
			}
			continue
		}

		if currentSection != nil {
			currentSection.Content += line + "\n"
		} else {
			basePrompt.WriteString(line)
			basePrompt.WriteString("\n")
		}
	}

	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	return &ProgressivePrompt{
		Sections:   sections,
		BasePrompt: basePrompt.String(),
	}, scanner.Err()
}

func (p *ProgressivePrompt) Base() string {
	return p.BasePrompt
}

func (p *ProgressivePrompt) WithSection(trigger string) string {
	var result strings.Builder
	result.WriteString(p.BasePrompt)
	result.WriteString("\n\n")

	for _, section := range p.Sections {
		if section.Trigger == trigger {
			result.WriteString(strings.TrimSpace(section.Content))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (p *ProgressivePrompt) RequiredSections() []string {
	var required []string
	for _, section := range p.Sections {
		if section.Required {
			required = append(required, section.Trigger)
		}
	}
	return required
}

func (p *ProgressivePrompt) AllSections() []string {
	var all []string
	for _, section := range p.Sections {
		all = append(all, section.Trigger)
	}
	return all
}

func (p *ProgressivePrompt) WithAllSections() string {
	var result strings.Builder
	result.WriteString(p.BasePrompt)
	result.WriteString("\n\n")

	for _, section := range p.Sections {
		result.WriteString(fmt.Sprintf("<!-- @section: %s -->\n", section.Trigger))
		result.WriteString(strings.TrimSpace(section.Content))
		result.WriteString("\n\n")
	}

	return result.String()
}
