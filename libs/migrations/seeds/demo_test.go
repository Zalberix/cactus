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
		"'Версия 2'",
		"'Старт системы'",
		"'Генерация письма'",
		"'Проверка email'",
		"'Письмо админа'",
		"'Отправка письма'",
		"'Итоговое письмо'",
	)

	mustNotContainAny(t, source,
		"'Demo Email Notification'",
		"'Email'",
		"'Summary email'",
		"'Отправить email'",
		"'Отправить Telegram'",
	)
}

func TestDemoSeedDefinesCurrentMapsVersionsAndSettings(t *testing.T) {
	source := readDemoSeedSource(t)

	mustContainAll(t, source,
		"locked_at",
		"published_at",
		"(1::int, 'Версия 1'::text, true::boolean, true::boolean",
		"(2::int, 'Версия 2'::text, true::boolean, true::boolean",
		"'2026-06-08 19:05:32.220064'::timestamp",
		"'2026-06-08 19:12:42.321461'::timestamp",
		"4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464",
		"0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541",
		"info@tyumen-city.ru",
		`'{"x":360,"y":200}'::jsonb`,
		`'{"x":600,"y":200}'::jsonb`,
		`'{"x":900,"y":60}'::jsonb`,
		`'{"x":900,"y":320}'::jsonb`,
		`'{"x":640,"y":200}'::jsonb`,
		"jsonb_build_object('target', 'body', 'source', format('$.steps.%s.output.body'",
		"'true'::text",
		"'false'::text",
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
		"'condition'",
		`'{"trigger":"system_message"}'::jsonb`,
		`'{"x":80,"y":200}'::jsonb`,
		"'Генерация письма'::text",
		"'Проверка email'::text",
		"'Письмо админа'::text",
		"'Отправка письма'::text",
		"'Итоговое письмо'::text",
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
