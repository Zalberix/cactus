package build

import (
	"github.com/pterm/pterm"
)

// LibsCmd builds all JS libraries with npm.
// No-op until frontend (Nuxt.js) is added in Phase 5.
func LibsCmd(_ string) error {
	pterm.Info.Println("No frontend libs to build (Phase 5)")
	return nil
}
