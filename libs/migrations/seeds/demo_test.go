package seeds

import (
	"os"
	"strings"
	"testing"
)

func TestDemoSeedDefinesRussianValidWorkflowVersion(t *testing.T) {
	source := readDemoSeedSource(t)

	mustContainAll(t, source,
		"'Уведомление о заявке'",
		"'Версия 1'",
		"'Старт системы'",
		"'Отправить email'",
		"'Отправить Telegram'",
		"'Итоговое письмо'",
	)

	mustNotContainAny(t, source,
		"'Demo Email Notification'",
		"'Email'",
		"'Summary email'",
	)
}

func TestDemoSeedDefinesStartStepAndCanvasPositions(t *testing.T) {
	source := readDemoSeedSource(t)

	mustContainAll(t, source,
		"control_kind",
		"control_settings",
		"canvas_position",
		"'control'",
		"'start'",
		`'{"trigger":"system_message"}'::jsonb`,
		`'{"x":80,"y":200}'::jsonb`,
		`'{"x":360,"y":80}'::jsonb`,
		`'{"x":360,"y":320}'::jsonb`,
		`'{"x":680,"y":200}'::jsonb`,
		"start_step",
		"email_step",
		"telegram_step",
		"summary_step",
	)
}

func readDemoSeedSource(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("demo.go")
	if err != nil {
		t.Fatalf("read demo seed: %v", err)
	}
	return string(raw)
}

func mustContainAll(t *testing.T, source string, values ...string) {
	t.Helper()

	for _, value := range values {
		if !strings.Contains(source, value) {
			t.Fatalf("demo seed does not contain %q", value)
		}
	}
}

func mustNotContainAny(t *testing.T, source string, values ...string) {
	t.Helper()

	for _, value := range values {
		if strings.Contains(source, value) {
			t.Fatalf("demo seed still contains %q", value)
		}
	}
}
