package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedDemo создаёт полный демо-стенд для тестирования POST /api/v1/messages/send.
//
// Зависимости (должны быть засеяны раньше):
//   - organization (code='test')
//   - admin user (admin@test.local)
//   - system ('Тестовая система')
//   - work_types (smtp, telegram)
//
// Создаёт:
//   - System token (pub: demo-public-token, priv: demo-private-token)
//   - Workflow "Demo Email Notification" с input_schema (JSON Schema)
//   - Активная версия workflow
//   - 3 шага: два независимых (smtp, telegram) + один зависимый
//   - Зависимости между шагами (DAG)
//   - Привязка system token → workflow (workflow_token)
func SeedDemo(ctx context.Context, db *pgxpool.Pool) error {
	// 1. System token
	// private_token хранится как SHA256 hex от "demo-private-token"
	_, err := db.Exec(ctx, `
		INSERT INTO system_token (system_id, name, public_token, private_token, is_active)
		VALUES (
			(SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1),
			'demo-token',
			'demo-public-token',
			encode(sha256('demo-private-token'::bytea), 'hex'),
			true
		)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed system token: %w", err)
	}

	// 2. Workflow с JSON Schema валидацией
	_, err = db.Exec(ctx, `
		INSERT INTO workflow (system_id, name, priority, description, input_schema)
		VALUES (
			(SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1),
			'Demo Email Notification',
			1,
			'Демо workflow: отправка email и telegram уведомлений',
			'{
				"type": "object",
				"properties": {
					"to": {"type": "string", "format": "email", "required": true},
					"subject": {"type": "string", "minLength": 1, "maxLength": 200, "required": true},
					"body": {"type": "string", "required": true},
					"telegram_chat_id": {"type": "string", "required": true}
				}
			}'::jsonb
		)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed workflow: %w", err)
	}

	// 3. Активная версия
	_, err = db.Exec(ctx, `
		INSERT INTO workflow_version (
			workflow_id, created_by_user_id, version_number, is_valid, is_active, traffic_weight, is_control_group
		)
		VALUES (
			(SELECT id FROM workflow WHERE name = 'Demo Email Notification' ORDER BY id LIMIT 1),
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			1,
			true,
			true,
			100,
			false
		)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed workflow version: %w", err)
	}

	// 4. Схемы настроек работников (Схема_настроек_работников) для smtp и telegram
	_, err = db.Exec(ctx, `
		INSERT INTO worker_settings_schema (work_type_id, version, settings_schema, input_schema, output_schema)
		VALUES
			(
				(SELECT id FROM work_type WHERE code = 'smtp'),
				'v1',
				'{"type": "object", "properties": {"host": {"type": "string", "required": true}, "port": {"type": "integer", "required": true}}}'::jsonb,
				'{
					"type": "object",
					"properties": {
						"to": {"type": "string", "required": true},
						"subject": {"type": "string", "required": true},
						"body": {"type": "string"}
					}
				}'::jsonb,
				'{"type": "object", "properties": {"message_id": {"type": "string"}}}'::jsonb
			),
			(
				(SELECT id FROM work_type WHERE code = 'telegram'),
				'v1',
				'{"type": "object", "properties": {"bot_token": {"type": "string", "required": true}}}'::jsonb,
				'{"type": "object", "properties": {"chat_id": {"type": "string", "required": true}, "text": {"type": "string", "required": true}}}'::jsonb,
				'{"type": "object", "properties": {"ok": {"type": "boolean"}}}'::jsonb
			)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed worker settings schema: %w", err)
	}

	// 5. Ревизии настроек работников (Ревизия_настроек_работников)
	_, err = db.Exec(ctx, `
		INSERT INTO worker_settings_revision (worker_settings_schema_id, created_by_user_id, settings_data)
		VALUES
			(
				(SELECT wss.id FROM worker_settings_schema wss
				 JOIN work_type wt ON wt.id = wss.work_type_id
				 WHERE wt.code = 'smtp' LIMIT 1),
				(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
				'{"host": "mailhog", "port": 1025}'::jsonb
			),
			(
				(SELECT wss.id FROM worker_settings_schema wss
				 JOIN work_type wt ON wt.id = wss.work_type_id
				 WHERE wt.code = 'telegram' LIMIT 1),
				(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
				'{"bot_token": "demo-bot-token"}'::jsonb
			)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed worker settings revision: %w", err)
	}

	// 6. Шаги workflow (DAG: step_email и step_telegram — независимые корни, step_summary — зависит от обоих)
	//
	// Визуально:
	//   [step_email]  ──success──┐
	//                            ├──► [step_summary]
	//   [step_telegram] ─success─┘
	//
	// CHECK constraint: task шаг требует work_type_id AND worker_settings_revision_id NOT NULL
	_, err = db.Exec(ctx, `
		INSERT INTO workflow_step (workflow_version_id, "name", step_type, work_type_id, worker_settings_revision_id, input_mapping)
		VALUES
			-- Step 1: Email (корень)
			(
				(SELECT wv.id FROM workflow_version wv JOIN workflow w ON w.id = wv.workflow_id
				 WHERE w.name = 'Demo Email Notification' AND wv.is_active = true LIMIT 1),
				'Email',
				'task',
				(SELECT id FROM work_type WHERE code = 'smtp'),
				(SELECT wsr.id FROM worker_settings_revision wsr
				 JOIN worker_settings_schema wss ON wss.id = wsr.worker_settings_schema_id
				 JOIN work_type wt ON wt.id = wss.work_type_id
				 WHERE wt.code = 'smtp' LIMIT 1),
				'[
					{"target": "to", "source": "$.message.value.to"},
					{"target": "subject", "source": "$.message.value.subject"},
					{"target": "body", "source": "$.message.value.body"}
				]'::jsonb
			),
			-- Step 2: Telegram (корень)
			(
				(SELECT wv.id FROM workflow_version wv JOIN workflow w ON w.id = wv.workflow_id
				 WHERE w.name = 'Demo Email Notification' AND wv.is_active = true LIMIT 1),
				'Telegram',
				'task',
				(SELECT id FROM work_type WHERE code = 'telegram'),
				(SELECT wsr.id FROM worker_settings_revision wsr
				 JOIN worker_settings_schema wss ON wss.id = wsr.worker_settings_schema_id
				 JOIN work_type wt ON wt.id = wss.work_type_id
				 WHERE wt.code = 'telegram' LIMIT 1),
				'[
					{"target": "chat_id", "source": "$.message.value.telegram_chat_id"},
					{"target": "text", "source": "$.message.value.body"}
				]'::jsonb
			),
			-- Step 3: Summary email (зависит от Step 1 и Step 2)
			(
				(SELECT wv.id FROM workflow_version wv JOIN workflow w ON w.id = wv.workflow_id
				 WHERE w.name = 'Demo Email Notification' AND wv.is_active = true LIMIT 1),
				'Summary email',
				'task',
				(SELECT id FROM work_type WHERE code = 'smtp'),
				(SELECT wsr.id FROM worker_settings_revision wsr
				 JOIN worker_settings_schema wss ON wss.id = wsr.worker_settings_schema_id
				 JOIN work_type wt ON wt.id = wss.work_type_id
				 WHERE wt.code = 'smtp' LIMIT 1),
				'[
					{"target": "to", "source": "$.message.value.to"},
					{"target": "subject", "source": "$.message.value.subject"}
				]'::jsonb
			)
	`)
	if err != nil {
		return fmt.Errorf("seed workflow steps: %w", err)
	}

	// 7. Зависимости: step_summary зависит от step_email и step_telegram
	_, err = db.Exec(ctx, `
		WITH ver AS (
			SELECT wv.id AS version_id
			FROM workflow_version wv
			JOIN workflow w ON w.id = wv.workflow_id
			WHERE w.name = 'Demo Email Notification' AND wv.is_active = true
			LIMIT 1
		),
		steps AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS rn
			FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			AND deleted_at IS NULL
		)
		INSERT INTO workflow_step_dependency (step_id, depends_on_step_id, outcome)
		VALUES
			-- step_summary (rn=3) зависит от step_email (rn=1) с outcome 'success'
			((SELECT id FROM steps WHERE rn = 3), (SELECT id FROM steps WHERE rn = 1), 'success'),
			-- step_summary (rn=3) зависит от step_telegram (rn=2) с outcome 'success'
			((SELECT id FROM steps WHERE rn = 3), (SELECT id FROM steps WHERE rn = 2), 'success')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed workflow step dependencies: %w", err)
	}

	// 8. Привязка system token к workflow
	_, err = db.Exec(ctx, `
		INSERT INTO workflow_token (system_token_id, workflow_id)
		VALUES (
			(SELECT id FROM system_token WHERE public_token = 'demo-public-token' ORDER BY id LIMIT 1),
			(SELECT id FROM workflow WHERE name = 'Demo Email Notification' ORDER BY id LIMIT 1)
		)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed workflow token: %w", err)
	}

	return nil
}
