package config

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
)

func (b *BrainsConfig) GetPersonaInstructions(persona string) string {
	personaText, found := b.Personas[persona]
	if !found {
		return ""
	}
	pterm.Debug.Printfln("user electing to leverage persona (%s) with text: %s", persona, personaText)
	return fmt.Sprintf("Human: %s\n\n", personaText)
}

func (b *BrainsConfig) GetConfig() *BrainsConfig { return b }

func (b *BrainsConfig) PreCommandsHook() error {
	for _, preCommand := range b.PreCommands {
		pterm.Info.Printfln("running command as part of pre_commands sequence %s", preCommand)
		cmd := exec.Command("bash", "-c", preCommand) // #nosec G204 -- preCommand is controlled intentionally and is variadic by design
		out, err := cmd.CombinedOutput()
		if err != nil {
			pterm.Error.Printf("error executing precommand (%s) in: %v\noutput: %s\n", preCommand, err, out)
			return err
		}
	}
	return nil
}
