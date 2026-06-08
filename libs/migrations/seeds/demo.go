package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedDemo создаёт демо-стенд для проверки отправки сообщений и редактора DAG.
//
// Зависимости должны быть засеяны раньше:
//   - organization (code='test')
//   - admin user (admin@test.local)
//   - system ('Тестовая система')
//   - work_types (smtp, telegram)
func SeedDemo(ctx context.Context, db *pgxpool.Pool) error {
	if err := seedDemoSystemToken(ctx, db); err != nil {
		return err
	}
	if err := seedDemoWorkflow(ctx, db); err != nil {
		return err
	}
	if err := seedDemoVersion(ctx, db); err != nil {
		return err
	}
	if err := seedDemoInputSchema(ctx, db); err != nil {
		return err
	}
	if err := seedDemoWorkerSettings(ctx, db); err != nil {
		return err
	}
	if err := seedDemoWorkflowSteps(ctx, db); err != nil {
		return err
	}
	if err := seedDemoWorkflowToken(ctx, db); err != nil {
		return err
	}
	return nil
}

func seedDemoSystemToken(ctx context.Context, db *pgxpool.Pool) error {
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
	return nil
}

func seedDemoWorkflow(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		INSERT INTO workflow (system_id, name, priority, description)
		SELECT
			(SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1),
			'Уведомление о заявке',
			1,
			'Демо-процесс: отправка email и Telegram-уведомлений по входящей заявке'
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow
			WHERE system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND name = 'Уведомление о заявке'
			  AND deleted_at IS NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("seed workflow: %w", err)
	}
	return nil
}

func seedDemoVersion(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		)
		INSERT INTO workflow_version (
			workflow_id, created_by_user_id, version_number, name, is_valid, is_active
		)
		SELECT
			(SELECT id FROM wf),
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			1,
			'Версия 1',
			true,
			true
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number = 1
			  AND deleted_at IS NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("seed workflow version: %w", err)
	}

	_, err = db.Exec(ctx, `
		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		)
		UPDATE workflow_version
		SET name = 'Версия 1',
		    is_valid = true,
		    is_active = true,
		    updated_at = CURRENT_TIMESTAMP
		WHERE workflow_id = (SELECT id FROM wf)
		  AND version_number = 1
		  AND deleted_at IS NULL
	`)
	if err != nil {
		return fmt.Errorf("update workflow version seed state: %w", err)
	}
	return nil
}

func seedDemoInputSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		ver AS (
			SELECT id, workflow_id
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number = 1
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		inserted_schema AS (
			INSERT INTO workflow_input_schema (workflow_id, code, version_number, schema_json, status, is_default)
			SELECT
				(SELECT id FROM wf),
				'v1',
				1,
				'{
					"type": "object",
					"properties": {
						"to": {"type": "string", "format": "email", "required": true},
						"subject": {"type": "string", "minLength": 1, "maxLength": 200, "required": true},
						"body": {"type": "string", "required": true},
						"telegram_chat_id": {"type": "string", "required": true}
					}
				}'::jsonb,
				'active',
				true
			WHERE NOT EXISTS (
				SELECT 1
				FROM workflow_input_schema wis
				WHERE wis.workflow_id = (SELECT id FROM wf)
				  AND wis.version_number = 1
				  AND wis.deleted_at IS NULL
			)
			RETURNING id
		),
		schema_row AS (
			SELECT id FROM inserted_schema
			UNION ALL
			SELECT wis.id
			FROM workflow_input_schema wis
			WHERE wis.workflow_id = (SELECT id FROM wf)
			  AND wis.version_number = 1
			  AND wis.deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		)
		INSERT INTO workflow_version_input_schema_compatibility (
			workflow_version_id,
			workflow_input_schema_id,
			compatibility_type,
			default_values,
			is_active,
			is_default_route
		)
		SELECT
			(SELECT id FROM ver),
			(SELECT id FROM schema_row),
			'native',
			'{}'::jsonb,
			true,
			true
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_version_input_schema_compatibility c
			WHERE c.workflow_version_id = (SELECT id FROM ver)
			  AND c.workflow_input_schema_id = (SELECT id FROM schema_row)
			  AND c.deleted_at IS NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("seed workflow input schema: %w", err)
	}
	return nil
}

func seedDemoWorkerSettings(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		WITH seed AS (
			SELECT *
			FROM (VALUES
				(
					'smtp'::text,
					'v1'::text,
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
					'telegram'::text,
					'v1'::text,
					'{"type": "object", "properties": {"bot_token": {"type": "string", "required": true}}}'::jsonb,
					'{"type": "object", "properties": {"chat_id": {"type": "string", "required": true}, "text": {"type": "string", "required": true}}}'::jsonb,
					'{"type": "object", "properties": {"ok": {"type": "boolean"}}}'::jsonb
				)
			) AS s(work_type_code, version, settings_schema, input_schema, output_schema)
		)
		INSERT INTO worker_settings_schema (work_type_id, version, settings_schema, input_schema, output_schema)
		SELECT
			(SELECT id FROM work_type WHERE code = seed.work_type_code),
			seed.version,
			seed.settings_schema,
			seed.input_schema,
			seed.output_schema
		FROM seed
		WHERE NOT EXISTS (
			SELECT 1
			FROM worker_settings_schema wss
			WHERE wss.work_type_id = (SELECT id FROM work_type WHERE code = seed.work_type_code)
			  AND wss.version = seed.version
			  AND wss.deleted_at IS NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("seed worker settings schema: %w", err)
	}

	_, err = db.Exec(ctx, `
		WITH seed AS (
			SELECT *
			FROM (VALUES
				('smtp'::text, '{"host": "mailhog", "port": 1025}'::jsonb),
				('telegram'::text, '{"bot_token": "demo-bot-token"}'::jsonb)
			) AS s(work_type_code, settings_data)
		)
		INSERT INTO worker_settings_revision (worker_settings_schema_id, created_by_user_id, settings_data)
		SELECT
			wss.id,
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			seed.settings_data
		FROM seed
		JOIN work_type wt ON wt.code = seed.work_type_code
		JOIN worker_settings_schema wss ON wss.work_type_id = wt.id AND wss.version = 'v1' AND wss.deleted_at IS NULL
		WHERE NOT EXISTS (
			SELECT 1
			FROM worker_settings_revision wsr
			WHERE wsr.worker_settings_schema_id = wss.id
			  AND wsr.settings_data = seed.settings_data
		)
	`)
	if err != nil {
		return fmt.Errorf("seed worker settings revision: %w", err)
	}
	return nil
}

func seedDemoWorkflowSteps(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		ver AS (
			SELECT id AS version_id
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number = 1
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(
					'Старт системы'::text,
					'control'::text,
					NULL::text,
					'start'::text,
					'{"trigger":"system_message"}'::jsonb,
					'[]'::jsonb,
					'{"x":80,"y":200}'::jsonb
				),
				(
					'Отправить email'::text,
					'task'::text,
					'smtp'::text,
					NULL::text,
					NULL::jsonb,
					'[
						{"target": "to", "source": "$.message.value.to"},
						{"target": "subject", "source": "$.message.value.subject"},
						{"target": "body", "source": "$.message.value.body"}
					]'::jsonb,
					'{"x":360,"y":80}'::jsonb
				),
				(
					'Отправить Telegram'::text,
					'task'::text,
					'telegram'::text,
					NULL::text,
					NULL::jsonb,
					'[
						{"target": "chat_id", "source": "$.message.value.telegram_chat_id"},
						{"target": "text", "source": "$.message.value.body"}
					]'::jsonb,
					'{"x":360,"y":320}'::jsonb
				),
				(
					'Итоговое письмо'::text,
					'task'::text,
					'smtp'::text,
					NULL::text,
					NULL::jsonb,
					'[
						{"target": "to", "source": "$.message.value.to"},
						{"target": "subject", "source": "$.message.value.subject"}
					]'::jsonb,
					'{"x":680,"y":200}'::jsonb
				)
			) AS s(name, step_type, work_type_code, control_kind, control_settings, input_mapping, canvas_position)
		)
		INSERT INTO workflow_step (
			workflow_version_id,
			name,
			step_type,
			work_type_id,
			worker_settings_revision_id,
			control_kind,
			control_settings,
			input_mapping,
			canvas_position
		)
		SELECT
			(SELECT version_id FROM ver),
			seed.name,
			seed.step_type,
			CASE
				WHEN seed.work_type_code IS NULL THEN NULL
				ELSE (SELECT id FROM work_type WHERE code = seed.work_type_code)
			END,
			CASE
				WHEN seed.work_type_code IS NULL THEN NULL
				ELSE (
					SELECT wsr.id
					FROM worker_settings_revision wsr
					JOIN worker_settings_schema wss ON wss.id = wsr.worker_settings_schema_id
					JOIN work_type wt ON wt.id = wss.work_type_id
					WHERE wt.code = seed.work_type_code
					  AND wss.version = 'v1'
					  AND wss.deleted_at IS NULL
					ORDER BY wsr.id DESC
					LIMIT 1
				)
			END,
			seed.control_kind,
			seed.control_settings,
			seed.input_mapping,
			seed.canvas_position
		FROM seed
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_step ws
			WHERE ws.workflow_version_id = (SELECT version_id FROM ver)
			  AND ws.name = seed.name
			  AND ws.deleted_at IS NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("seed workflow steps: %w", err)
	}

	_, err = db.Exec(ctx, `
		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		ver AS (
			SELECT id AS version_id
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number = 1
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		start_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Старт системы'
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		email_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Отправить email'
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		telegram_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Отправить Telegram'
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		summary_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Итоговое письмо'
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		edges AS (
			SELECT (SELECT id FROM email_step) AS step_id, (SELECT id FROM start_step) AS depends_on_step_id, 'success'::text AS outcome
			UNION ALL
			SELECT (SELECT id FROM telegram_step), (SELECT id FROM start_step), 'success'::text
			UNION ALL
			SELECT (SELECT id FROM summary_step), (SELECT id FROM email_step), 'success'::text
			UNION ALL
			SELECT (SELECT id FROM summary_step), (SELECT id FROM telegram_step), 'success'::text
		)
		INSERT INTO workflow_step_dependency (step_id, depends_on_step_id, outcome)
		SELECT step_id, depends_on_step_id, outcome
		FROM edges
		WHERE step_id IS NOT NULL
		  AND depends_on_step_id IS NOT NULL
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed workflow step dependencies: %w", err)
	}
	return nil
}

func seedDemoWorkflowToken(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		INSERT INTO workflow_token (system_token_id, workflow_id)
		SELECT
			(SELECT id FROM system_token WHERE public_token = 'demo-public-token' ORDER BY id LIMIT 1),
			(SELECT id FROM workflow WHERE name = 'Уведомление о заявке' ORDER BY id LIMIT 1)
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_token wt
			WHERE wt.system_token_id = (SELECT id FROM system_token WHERE public_token = 'demo-public-token' ORDER BY id LIMIT 1)
			  AND wt.workflow_id = (SELECT id FROM workflow WHERE name = 'Уведомление о заявке' ORDER BY id LIMIT 1)
		)
	`)
	if err != nil {
		return fmt.Errorf("seed workflow token: %w", err)
	}
	return nil
}
