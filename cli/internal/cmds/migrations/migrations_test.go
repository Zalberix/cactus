package migrations

import "testing"

func TestCommandSupportsPlanPluralAlias(t *testing.T) {
	cmd := Command()
	for _, alias := range cmd.Aliases {
		if alias == "migrations" {
			return
		}
	}
	t.Fatalf("expected migration command to support migrations alias, aliases=%v", cmd.Aliases)
}

func TestStatusCommandDeclaresMigrationsPathFlag(t *testing.T) {
	for _, flag := range StatusMigrationCmd.Flags {
		if flag.Names()[0] == "migrations-path" {
			return
		}
	}
	t.Fatalf("expected status command to declare migrations-path flag")
}
