package prompt

import (
	"strings"

	"github.com/manifoldco/promptui"
)

type Prompt interface {
	AskConfirm(string) bool
}

type RealPrompt struct{}

// NewRealPrompt is the factory builder for RealPrompt
func NewRealPrompt() *RealPrompt {
	return &RealPrompt{}
}

// AskConfirm will use promptui to provide a confirmation
func (r *RealPrompt) AskConfirm(confirm string) bool {
	p := promptui.Prompt{
		Label:     confirm,
		IsConfirm: true,
	}
	if res, err := p.Run(); err != nil {
		return false
	} else {
		switch strings.ToLower(res) {
		case "y":
			return true
		case "yes":
			return true
		default:
			return false
		}
	}
}

// Mock Prompt that always returns true
type FakePromptYes struct{}

func NewFakePromptYes() *FakePromptYes {
	return &FakePromptYes{}
}

func (f FakePromptYes) AskConfirm(_ string) bool {
	return true
}

// Mock Prompt that always returns false
type FakePromptNo struct{}

func NewFakePromptNo() *FakePromptNo {
	return &FakePromptNo{}
}

func (f FakePromptNo) AskConfirm(_ string) bool {
	return false
}
