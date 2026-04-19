package cli

import (
	"errors"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

type inputPrompter func(message string, defaultValue string) (string, error)
type selectPrompter func(message string, choices []string, defaultIndex int) (string, error)
type confirmPrompter func(message string, defaultValue bool) (bool, error)

type promptUI struct {
	inputPrompter   inputPrompter
	selectPrompter  selectPrompter
	confirmPrompter confirmPrompter
}

func newPromptUI() *promptUI {
	ui := &promptUI{}
	ui.inputPrompter = ui.runInputPrompt
	ui.selectPrompter = ui.runSelectPrompt
	ui.confirmPrompter = ui.runConfirmPrompt
	return ui
}

func (ui *promptUI) askInput(message string, defaultValue string) (string, error) {
	return ui.inputPrompter(message, defaultValue)
}

func (ui *promptUI) askSelect(message string, choices []string, defaultIndex int) (string, error) {
	return ui.selectPrompter(message, choices, defaultIndex)
}

func (ui *promptUI) askConfirm(message string, defaultValue bool) (bool, error) {
	return ui.confirmPrompter(message, defaultValue)
}

func (ui *promptUI) runInputPrompt(message string, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:     message,
		Default:   defaultValue,
		AllowEdit: true,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}
	return prompt.Run()
}

func (ui *promptUI) runSelectPrompt(message string, choices []string, defaultIndex int) (string, error) {
	prompt := promptui.Select{
		Label:  message,
		Items:  choices,
		Size:   len(choices),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}
	_, result, err := prompt.RunCursorAt(defaultIndex, 0)
	return result, err
}

func (ui *promptUI) runConfirmPrompt(message string, defaultValue bool) (bool, error) {
	defaultAnswer := "n"
	if defaultValue {
		defaultAnswer = "y"
	}

	prompt := promptui.Prompt{
		Label:     message,
		Default:   defaultAnswer,
		IsConfirm: true,
		AllowEdit: true,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}

	result, err := prompt.Run()
	return runConfirmPromptResult(result, err, defaultValue)
}

func runConfirmPromptResult(result string, err error, defaultValue bool) (bool, error) {
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false, nil
		}
		return false, err
	}

	switch strings.ToLower(result) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return defaultValue, nil
	}
}
