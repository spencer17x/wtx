package cli

import (
	"errors"
	"testing"

	"github.com/manifoldco/promptui"
)

func TestPromptUIAskSelectUsesInjectedPrompter(t *testing.T) {
	t.Parallel()

	ui := &promptUI{
		selectPrompter: func(message string, choices []string, defaultIndex int) (string, error) {
			if message != "How should the worktree branch be created?" {
				t.Fatalf("message = %q", message)
			}
			if defaultIndex != 0 {
				t.Fatalf("defaultIndex = %d, want 0", defaultIndex)
			}
			if len(choices) != 2 || choices[1] != "Create a new branch" {
				t.Fatalf("choices = %#v", choices)
			}
			return choices[1], nil
		},
	}

	got, err := ui.askSelect(
		"How should the worktree branch be created?",
		[]string{"Use an existing branch", "Create a new branch"},
		0,
	)
	if err != nil {
		t.Fatalf("askSelect: %v", err)
	}
	if got != "Create a new branch" {
		t.Fatalf("got %q, want %q", got, "Create a new branch")
	}
}

func TestPromptUIAskInputUsesInjectedPrompter(t *testing.T) {
	t.Parallel()

	ui := &promptUI{
		inputPrompter: func(message string, defaultValue string) (string, error) {
			if message != "Existing branch name" {
				t.Fatalf("message = %q", message)
			}
			if defaultValue != "main" {
				t.Fatalf("defaultValue = %q, want %q", defaultValue, "main")
			}
			return "feature/test", nil
		},
	}

	got, err := ui.askInput("Existing branch name", "main")
	if err != nil {
		t.Fatalf("askInput: %v", err)
	}
	if got != "feature/test" {
		t.Fatalf("got %q, want %q", got, "feature/test")
	}
}

func TestPromptUIAskConfirmUsesInjectedPrompter(t *testing.T) {
	t.Parallel()

	ui := &promptUI{
		confirmPrompter: func(message string, defaultValue bool) (bool, error) {
			if message != "Create the worktree with this plan?" {
				t.Fatalf("message = %q", message)
			}
			if !defaultValue {
				t.Fatalf("defaultValue = %v, want true", defaultValue)
			}
			return true, nil
		},
	}

	got, err := ui.askConfirm("Create the worktree with this plan?", true)
	if err != nil {
		t.Fatalf("askConfirm: %v", err)
	}
	if !got {
		t.Fatalf("got %v, want true", got)
	}
}

func TestPromptUIAskConfirmReturnsInjectedError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("interrupted")
	ui := &promptUI{
		confirmPrompter: func(message string, defaultValue bool) (bool, error) {
			return false, wantErr
		},
	}

	_, err := ui.askConfirm("Create the worktree with this plan?", true)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestNewPromptUIInitializesPromptFunctions(t *testing.T) {
	t.Parallel()

	ui := newPromptUI()
	if ui == nil {
		t.Fatal("ui = nil")
	}
	if ui.inputPrompter == nil {
		t.Fatal("inputPrompter = nil")
	}
	if ui.selectPrompter == nil {
		t.Fatal("selectPrompter = nil")
	}
	if ui.confirmPrompter == nil {
		t.Fatal("confirmPrompter = nil")
	}
}

func TestRunConfirmPromptResultReturnsFalseOnPromptUIAbort(t *testing.T) {
	t.Parallel()

	got, err := runConfirmPromptResult("", promptui.ErrAbort, true)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got {
		t.Fatalf("got %v, want false", got)
	}
}
