package config

import (
	"fmt"

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
