package worker

import "fmt"

type Variant struct {
	Name     string
	Manifest ManifestSpec
	Handler  TaskHandler
}

func SelectVariant(variants map[string]Variant, name string) (Variant, error) {
	variant, ok := variants[name]
	if !ok {
		return Variant{}, fmt.Errorf("unknown worker variant %q", name)
	}
	if variant.Name == "" {
		variant.Name = name
	}
	return variant, nil
}
