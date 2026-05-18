package main

import (
	"testing"

	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/internal/configpub"
	"github.com/zalberix/cactus/apps/core/internal/store"
	"github.com/zalberix/cactus/libs/bus"
)

func TestConfigPublisherFxWiringUsesConcreteStoreAndBus(t *testing.T) {
	err := fx.ValidateApp(
		fx.Provide(
			func() *store.Store { return nil },
			func() *bus.Bus { return nil },
			newConfigPubService,
		),
		fx.Invoke(func(*configpub.Service) {}),
	)
	if err != nil {
		t.Fatalf("validate config publisher graph: %v", err)
	}
}
