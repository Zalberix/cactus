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
//   - work_types (smtp, telegram, html)
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
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(1::int, 'Версия 1'::text, true::boolean, true::boolean, '2026-06-08 19:05:32.220064'::timestamp, '2026-06-08 19:05:32.220064'::timestamp, NULL::timestamp),
				(2::int, 'Версия 2'::text, true::boolean, true::boolean, '2026-06-08 19:12:42.321461'::timestamp, '2026-06-08 19:12:42.321461'::timestamp, NULL::timestamp)
			) AS s(version_number, name, is_valid, is_active, locked_at, published_at, archived_at)
		)
		INSERT INTO workflow_version (
			workflow_id,
			created_by_user_id,
			updated_by_user_id,
			version_number,
			name,
			is_valid,
			is_active,
			locked_at,
			published_at,
			archived_at
		)
		SELECT
			(SELECT id FROM wf),
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			seed.version_number,
			seed.name,
			seed.is_valid,
			seed.is_active,
			seed.locked_at,
			seed.published_at,
			seed.archived_at
		FROM seed
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_version existing_version
			WHERE existing_version.workflow_id = (SELECT id FROM wf)
			  AND existing_version.version_number = seed.version_number
			  AND existing_version.deleted_at IS NULL
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
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(1::int, 'Версия 1'::text, true::boolean, true::boolean, '2026-06-08 19:05:32.220064'::timestamp, '2026-06-08 19:05:32.220064'::timestamp, NULL::timestamp),
				(2::int, 'Версия 2'::text, true::boolean, true::boolean, '2026-06-08 19:12:42.321461'::timestamp, '2026-06-08 19:12:42.321461'::timestamp, NULL::timestamp)
			) AS s(version_number, name, is_valid, is_active, locked_at, published_at, archived_at)
		)
		UPDATE workflow_version wv
		SET name = seed.name,
		    is_valid = seed.is_valid,
		    is_active = seed.is_active,
		    locked_at = seed.locked_at,
		    published_at = seed.published_at,
		    archived_at = seed.archived_at,
		    updated_by_user_id = (SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
		    updated_at = CURRENT_TIMESTAMP
		FROM seed
		WHERE wv.workflow_id = (SELECT id FROM wf)
		  AND wv.version_number = seed.version_number
		  AND wv.deleted_at IS NULL
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
		)
		UPDATE workflow_input_schema
		SET is_default = false,
		    updated_at = CURRENT_TIMESTAMP
		WHERE workflow_id = (SELECT id FROM wf)
		  AND deleted_at IS NULL;

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(
					'v1'::text,
					1::int,
					'{
						"type": "object",
						"properties": {
							"to": {"type": "string", "format": "email", "required": true},
							"subject": {"type": "string", "minLength": 1, "maxLength": 200, "required": true},
							"body": {"type": "string", "required": true},
							"fields": {"type": "object", "required": true, "description": "Template values as a JSON object"},
							"template": {"type": "string", "required": true, "description": "Template file name without path"}
						}
					}'::jsonb,
					'active'::text,
					true::boolean
				),
				(
					'v2'::text,
					2::int,
					'{
						"type": "object",
						"properties": {
							"to": {"type": "string", "format": "email", "required": true},
							"subject": {"type": "string", "minLength": 1, "maxLength": 200, "required": true},
							"body": {"type": "string", "required": true},
							"fields": {"type": "object", "required": true, "description": "Template values as a JSON object"},
							"template": {"type": "string", "required": true, "description": "Template file name without path"},
							"is_admin": {"type": "boolean", "required": true}
						}
					}'::jsonb,
					'active'::text,
					false::boolean
				)
			) AS s(code, version_number, schema_json, status, is_default)
		)
		INSERT INTO workflow_input_schema (
			workflow_id,
			code,
			version_number,
			schema_json,
			status,
			is_default,
			created_by_user_id,
			updated_by_user_id
		)
		SELECT
			(SELECT id FROM wf),
			seed.code,
			seed.version_number,
			seed.schema_json,
			seed.status,
			seed.is_default,
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1)
		FROM seed
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_input_schema wis
			WHERE wis.workflow_id = (SELECT id FROM wf)
			  AND wis.version_number = seed.version_number
			  AND wis.deleted_at IS NULL
		);

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		seed AS (
			SELECT *
			FROM (VALUES
				('v1'::text, 1::int, '{
					"type": "object",
					"properties": {
						"to": {"type": "string", "format": "email", "required": true},
						"subject": {"type": "string", "minLength": 1, "maxLength": 200, "required": true},
						"body": {"type": "string", "required": true},
						"fields": {"type": "object", "required": true, "description": "Template values as a JSON object"},
						"template": {"type": "string", "required": true, "description": "Template file name without path"}
					}
				}'::jsonb, 'active'::text, true::boolean),
				('v2'::text, 2::int, '{
					"type": "object",
					"properties": {
						"to": {"type": "string", "format": "email", "required": true},
						"subject": {"type": "string", "minLength": 1, "maxLength": 200, "required": true},
						"body": {"type": "string", "required": true},
						"fields": {"type": "object", "required": true, "description": "Template values as a JSON object"},
						"template": {"type": "string", "required": true, "description": "Template file name without path"},
						"is_admin": {"type": "boolean", "required": true}
					}
				}'::jsonb, 'active'::text, false::boolean)
			) AS s(code, version_number, schema_json, status, is_default)
		)
		UPDATE workflow_input_schema wis
		SET code = seed.code,
		    schema_json = seed.schema_json,
		    status = seed.status,
		    is_default = seed.is_default,
		    updated_by_user_id = (SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
		    updated_at = CURRENT_TIMESTAMP
		FROM seed
		WHERE wis.workflow_id = (SELECT id FROM wf)
		  AND wis.version_number = seed.version_number
		  AND wis.deleted_at IS NULL;

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(1::int, 1::int),
				(2::int, 2::int)
			) AS s(workflow_version_number, schema_version_number)
		)
		INSERT INTO workflow_version_input_schema_compatibility (
			workflow_version_id,
			workflow_input_schema_id,
			compatibility_type,
			workflow_input_mapper_id,
			default_values,
			is_active,
			is_default_route,
			created_by_user_id,
			updated_by_user_id
		)
		SELECT
			wv.id,
			wis.id,
			'native',
			NULL,
			'{}'::jsonb,
			true,
			true,
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1)
		FROM seed
		JOIN workflow_version wv ON wv.workflow_id = (SELECT id FROM wf)
			AND wv.version_number = seed.workflow_version_number
			AND wv.deleted_at IS NULL
		JOIN workflow_input_schema wis ON wis.workflow_id = (SELECT id FROM wf)
			AND wis.version_number = seed.schema_version_number
			AND wis.deleted_at IS NULL
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_version_input_schema_compatibility c
			WHERE c.workflow_version_id = wv.id
			  AND c.workflow_input_schema_id = wis.id
			  AND c.deleted_at IS NULL
		);

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(1::int, 1::int),
				(2::int, 2::int)
			) AS s(workflow_version_number, schema_version_number)
		)
		UPDATE workflow_version_input_schema_compatibility c
		SET compatibility_type = 'native',
		    workflow_input_mapper_id = NULL,
		    default_values = '{}'::jsonb,
		    is_active = true,
		    is_default_route = true,
		    updated_by_user_id = (SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
		    updated_at = CURRENT_TIMESTAMP
		FROM seed
		JOIN workflow_version wv ON wv.workflow_id = (SELECT id FROM wf)
			AND wv.version_number = seed.workflow_version_number
			AND wv.deleted_at IS NULL
		JOIN workflow_input_schema wis ON wis.workflow_id = (SELECT id FROM wf)
			AND wis.version_number = seed.schema_version_number
			AND wis.deleted_at IS NULL
		WHERE c.workflow_version_id = wv.id
		  AND c.workflow_input_schema_id = wis.id
		  AND c.deleted_at IS NULL
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
					'smtp'::text,
					'0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text,
					'{"type":"object","properties":{"from":{"type":"string","required":true},"host":{"type":"string","required":true},"port":{"type":"integer","required":true}}}'::jsonb,
					'{"type":"object","properties":{"to":{"type":"string","required":true},"subject":{"type":"string","required":true},"body":{"type":"string"}}}'::jsonb,
					'{"type":"object","properties":{"sent_at":{"type":"string"},"message_id":{"type":"string"},"recipients_count":{"type":"integer"}}}'::jsonb
				),
				(
					'smtp'::text,
					'8b206a607b285dfe131e06209716da0ecb36f1c085d2cbb8ac7c59c85ee78a70'::text,
					'{"type":"object","properties":{"host":{"type":"string","required":true},"port":{"type":"integer","required":true},"from":{"type":"string","required":true},"auth":{"type":"string","enum":["none","plain","login"]},"tls":{"type":"string","enum":["none","tls","starttls"]}}}'::jsonb,
					'{"type":"object","properties":{"to":{"type":"string","required":true},"subject":{"type":"string","required":true},"body":{"type":"string","required":true},"cc":{"type":"string"}}}'::jsonb,
					'{"type":"object","properties":{"sent_at":{"type":"string"},"message_id":{"type":"string"},"recipients_count":{"type":"integer"}}}'::jsonb
				),
				(
					'telegram'::text,
					'v1'::text,
					'{"type": "object", "properties": {"bot_token": {"type": "string", "required": true}}}'::jsonb,
					'{"type": "object", "properties": {"chat_id": {"type": "string", "required": true}, "text": {"type": "string", "required": true}}}'::jsonb,
					'{"type": "object", "properties": {"ok": {"type": "boolean"}}}'::jsonb
				),
				(
					'html'::text,
					'4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text,
					'{"type":"object","properties":{}}'::jsonb,
					'{"type":"object","properties":{"fields":{"type":"object","required":true,"description":"Template values as a JSON object"},"template":{"type":"string","required":true,"description":"Template file name without path"}}}'::jsonb,
					'{"type":"object","properties":{"body":{"type":"string","required":true,"description":"Rendered HTML body"}}}'::jsonb
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
				(
					'smtp'::text,
					'v1'::text,
					'{"type": "object", "properties": {"host": {"type": "string", "required": true}, "port": {"type": "integer", "required": true}}}'::jsonb,
					'{"type":"object","properties":{"to":{"type":"string","required":true},"subject":{"type":"string","required":true},"body":{"type":"string"}}}'::jsonb,
					'{"type": "object", "properties": {"message_id": {"type": "string"}}}'::jsonb
				),
				(
					'smtp'::text,
					'0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text,
					'{"type":"object","properties":{"from":{"type":"string","required":true},"host":{"type":"string","required":true},"port":{"type":"integer","required":true}}}'::jsonb,
					'{"type":"object","properties":{"to":{"type":"string","required":true},"subject":{"type":"string","required":true},"body":{"type":"string"}}}'::jsonb,
					'{"type":"object","properties":{"sent_at":{"type":"string"},"message_id":{"type":"string"},"recipients_count":{"type":"integer"}}}'::jsonb
				),
				(
					'smtp'::text,
					'8b206a607b285dfe131e06209716da0ecb36f1c085d2cbb8ac7c59c85ee78a70'::text,
					'{"type":"object","properties":{"host":{"type":"string","required":true},"port":{"type":"integer","required":true},"from":{"type":"string","required":true},"auth":{"type":"string","enum":["none","plain","login"]},"tls":{"type":"string","enum":["none","tls","starttls"]}}}'::jsonb,
					'{"type":"object","properties":{"to":{"type":"string","required":true},"subject":{"type":"string","required":true},"body":{"type":"string","required":true},"cc":{"type":"string"}}}'::jsonb,
					'{"type":"object","properties":{"sent_at":{"type":"string"},"message_id":{"type":"string"},"recipients_count":{"type":"integer"}}}'::jsonb
				),
				(
					'telegram'::text,
					'v1'::text,
					'{"type": "object", "properties": {"bot_token": {"type": "string", "required": true}}}'::jsonb,
					'{"type": "object", "properties": {"chat_id": {"type": "string", "required": true}, "text": {"type": "string", "required": true}}}'::jsonb,
					'{"type": "object", "properties": {"ok": {"type": "boolean"}}}'::jsonb
				),
				(
					'html'::text,
					'4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text,
					'{"type":"object","properties":{}}'::jsonb,
					'{"type":"object","properties":{"fields":{"type":"object","required":true,"description":"Template values as a JSON object"},"template":{"type":"string","required":true,"description":"Template file name without path"}}}'::jsonb,
					'{"type":"object","properties":{"body":{"type":"string","required":true,"description":"Rendered HTML body"}}}'::jsonb
				)
			) AS s(work_type_code, version, settings_schema, input_schema, output_schema)
		)
		UPDATE worker_settings_schema wss
		SET settings_schema = seed.settings_schema,
		    input_schema = seed.input_schema,
		    output_schema = seed.output_schema,
		    updated_at = CURRENT_TIMESTAMP
		FROM seed
		JOIN work_type wt ON wt.code = seed.work_type_code
		WHERE wss.work_type_id = wt.id
		  AND wss.version = seed.version
		  AND wss.deleted_at IS NULL;

		WITH seed AS (
			SELECT *
			FROM (VALUES
				('smtp'::text, 'v1'::text, '{"host": "mailhog", "port": 1025}'::jsonb),
				('smtp'::text, 'v1'::text, '{"host": "localhost", "port": 1025}'::jsonb),
				('smtp'::text, '0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text, '{}'::jsonb),
				('smtp'::text, '0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text, '{"from": "info@tyumen-city.ru", "host": "localhost", "port": 1025}'::jsonb),
				('smtp'::text, '8b206a607b285dfe131e06209716da0ecb36f1c085d2cbb8ac7c59c85ee78a70'::text, '{}'::jsonb),
				('telegram'::text, 'v1'::text, '{"bot_token": "demo-bot-token"}'::jsonb),
				('html'::text, '4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text, '{}'::jsonb)
			) AS s(work_type_code, schema_version, settings_data)
		)
		INSERT INTO worker_settings_revision (worker_settings_schema_id, created_by_user_id, settings_data)
		SELECT
			wss.id,
			(SELECT id FROM "user" WHERE email = 'admin@test.local' ORDER BY id LIMIT 1),
			seed.settings_data
		FROM seed
		JOIN work_type wt ON wt.code = seed.work_type_code
		JOIN worker_settings_schema wss ON wss.work_type_id = wt.id AND wss.version = seed.schema_version AND wss.deleted_at IS NULL
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
		versions AS (
			SELECT id AS version_id, version_number
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number IN (1, 2)
			  AND deleted_at IS NULL
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(
					1::int,
					'Старт системы'::text,
					'control'::text,
					NULL::text,
					NULL::text,
					NULL::jsonb,
					'start'::text,
					'{"trigger":"system_message"}'::jsonb,
					'[]'::jsonb,
					'{"x":80,"y":200}'::jsonb
				),
				(
					1::int,
					'Генерация письма'::text,
					'task'::text,
					'html'::text,
					'4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text,
					'{}'::jsonb,
					NULL::text,
					NULL::jsonb,
					'[
						{"target": "fields", "source": "$.message.value.fields"},
						{"target": "template", "source": "$.message.value.template"}
					]'::jsonb,
					'{"x":360,"y":200}'::jsonb
				),
				(
					1::int,
					'Итоговое письмо'::text,
					'task'::text,
					'smtp'::text,
					'v1'::text,
					'{"host": "mailhog", "port": 1025}'::jsonb,
					NULL::text,
					NULL::jsonb,
					'[]'::jsonb,
					'{"x":640,"y":200}'::jsonb
				),
				(
					2::int,
					'Старт системы'::text,
					'control'::text,
					NULL::text,
					NULL::text,
					NULL::jsonb,
					'start'::text,
					'{"trigger":"system_message"}'::jsonb,
					'[]'::jsonb,
					'{"x":80,"y":200}'::jsonb
				),
				(
					2::int,
					'Генерация письма'::text,
					'task'::text,
					'html'::text,
					'4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text,
					'{}'::jsonb,
					NULL::text,
					NULL::jsonb,
					'[
						{"target": "fields", "source": "$.message.value.fields"},
						{"target": "template", "source": "$.message.value.template"}
					]'::jsonb,
					'{"x":360,"y":200}'::jsonb
				),
				(
					2::int,
					'Проверка email'::text,
					'control'::text,
					NULL::text,
					NULL::text,
					NULL::jsonb,
					'condition'::text,
					'{"left":"$.message.value.to","right":"zalberix@gmail.com","operator":"eq"}'::jsonb,
					NULL::jsonb,
					'{"x":600,"y":200}'::jsonb
				),
				(
					2::int,
					'Письмо админа'::text,
					'task'::text,
					'smtp'::text,
					'0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text,
					'{"from": "info@tyumen-city.ru", "host": "localhost", "port": 1025}'::jsonb,
					NULL::text,
					NULL::jsonb,
					'[
						{"target": "to", "source": "$.message.value.to"},
						{"target": "subject", "source": "$.message.value.subject"},
						{"target": "body", "source": "Ты был выбран"}
					]'::jsonb,
					'{"x":900,"y":60}'::jsonb
				),
				(
					2::int,
					'Отправка письма'::text,
					'task'::text,
					'smtp'::text,
					'0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text,
					'{"from": "info@tyumen-city.ru", "host": "localhost", "port": 1025}'::jsonb,
					NULL::text,
					NULL::jsonb,
					'[]'::jsonb,
					'{"x":900,"y":320}'::jsonb
				)
			) AS s(version_number, name, step_type, work_type_code, settings_schema_version, settings_data, control_kind, control_settings, input_mapping, canvas_position)
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
			versions.version_id,
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
					  AND wss.version = seed.settings_schema_version
					  AND wsr.settings_data = seed.settings_data
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
		JOIN versions ON versions.version_number = seed.version_number
		WHERE NOT EXISTS (
			SELECT 1
			FROM workflow_step ws
			WHERE ws.workflow_version_id = versions.version_id
			  AND ws.name = seed.name
			  AND ws.deleted_at IS NULL
		);

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		versions AS (
			SELECT id AS version_id, version_number
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number IN (1, 2)
			  AND deleted_at IS NULL
		),
		seed AS (
			SELECT *
			FROM (VALUES
				(1::int, 'Старт системы'::text, 'control'::text, NULL::text, NULL::text, NULL::jsonb, 'start'::text, '{"trigger":"system_message"}'::jsonb, '[]'::jsonb, '{"x":80,"y":200}'::jsonb),
				(1::int, 'Генерация письма'::text, 'task'::text, 'html'::text, '4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text, '{}'::jsonb, NULL::text, NULL::jsonb, '[{"target":"fields","source":"$.message.value.fields"},{"target":"template","source":"$.message.value.template"}]'::jsonb, '{"x":360,"y":200}'::jsonb),
				(1::int, 'Итоговое письмо'::text, 'task'::text, 'smtp'::text, 'v1'::text, '{"host": "mailhog", "port": 1025}'::jsonb, NULL::text, NULL::jsonb, '[]'::jsonb, '{"x":640,"y":200}'::jsonb),
				(2::int, 'Старт системы'::text, 'control'::text, NULL::text, NULL::text, NULL::jsonb, 'start'::text, '{"trigger":"system_message"}'::jsonb, '[]'::jsonb, '{"x":80,"y":200}'::jsonb),
				(2::int, 'Генерация письма'::text, 'task'::text, 'html'::text, '4ab1225482a1a041a7d9a31f23a99112b20567cf94a4a41b5944d89243d43464'::text, '{}'::jsonb, NULL::text, NULL::jsonb, '[{"target":"fields","source":"$.message.value.fields"},{"target":"template","source":"$.message.value.template"}]'::jsonb, '{"x":360,"y":200}'::jsonb),
				(2::int, 'Проверка email'::text, 'control'::text, NULL::text, NULL::text, NULL::jsonb, 'condition'::text, '{"left":"$.message.value.to","right":"zalberix@gmail.com","operator":"eq"}'::jsonb, NULL::jsonb, '{"x":600,"y":200}'::jsonb),
				(2::int, 'Письмо админа'::text, 'task'::text, 'smtp'::text, '0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text, '{"from": "info@tyumen-city.ru", "host": "localhost", "port": 1025}'::jsonb, NULL::text, NULL::jsonb, '[{"target":"to","source":"$.message.value.to"},{"target":"subject","source":"$.message.value.subject"},{"target":"body","source":"Ты был выбран"}]'::jsonb, '{"x":900,"y":60}'::jsonb),
				(2::int, 'Отправка письма'::text, 'task'::text, 'smtp'::text, '0952a9a06c6fbd5136e9c05aa648bd1bd583d619ee568dbfc2172cad44c18541'::text, '{"from": "info@tyumen-city.ru", "host": "localhost", "port": 1025}'::jsonb, NULL::text, NULL::jsonb, '[]'::jsonb, '{"x":900,"y":320}'::jsonb)
			) AS s(version_number, name, step_type, work_type_code, settings_schema_version, settings_data, control_kind, control_settings, input_mapping, canvas_position)
		)
		UPDATE workflow_step ws
		SET step_type = seed.step_type,
		    work_type_id = CASE
			    WHEN seed.work_type_code IS NULL THEN NULL
			    ELSE (SELECT id FROM work_type WHERE code = seed.work_type_code)
		    END,
		    worker_settings_revision_id = CASE
			    WHEN seed.work_type_code IS NULL THEN NULL
			    ELSE (
				    SELECT wsr.id
				    FROM worker_settings_revision wsr
				    JOIN worker_settings_schema wss ON wss.id = wsr.worker_settings_schema_id
				    JOIN work_type wt ON wt.id = wss.work_type_id
				    WHERE wt.code = seed.work_type_code
				      AND wss.version = seed.settings_schema_version
				      AND wsr.settings_data = seed.settings_data
				      AND wss.deleted_at IS NULL
				    ORDER BY wsr.id DESC
				    LIMIT 1
			    )
		    END,
		    control_kind = seed.control_kind,
		    control_settings = seed.control_settings,
		    input_mapping = seed.input_mapping,
		    canvas_position = seed.canvas_position,
		    updated_at = CURRENT_TIMESTAMP
		FROM seed
		JOIN versions ON versions.version_number = seed.version_number
		WHERE ws.workflow_version_id = versions.version_id
		  AND ws.name = seed.name
		  AND ws.deleted_at IS NULL;

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		versions AS (
			SELECT id AS version_id, version_number
			FROM workflow_version
			WHERE workflow_id = (SELECT id FROM wf)
			  AND version_number IN (1, 2)
			  AND deleted_at IS NULL
		),
		seed_names AS (
			SELECT *
			FROM (VALUES
				(1::int, 'Старт системы'::text),
				(1::int, 'Генерация письма'::text),
				(1::int, 'Итоговое письмо'::text),
				(2::int, 'Старт системы'::text),
				(2::int, 'Генерация письма'::text),
				(2::int, 'Проверка email'::text),
				(2::int, 'Письмо админа'::text),
				(2::int, 'Отправка письма'::text)
			) AS s(version_number, name)
		)
		UPDATE workflow_step ws
		SET deleted_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		FROM versions
		WHERE ws.workflow_version_id = versions.version_id
		  AND ws.deleted_at IS NULL
		  AND NOT EXISTS (
			  SELECT 1
			  FROM seed_names
			  WHERE seed_names.version_number = versions.version_number
			    AND seed_names.name = ws.name
		  );

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
		generator_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Генерация письма'
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
		)
		UPDATE workflow_step ws
		SET input_mapping = jsonb_build_array(
			    jsonb_build_object('target', 'to', 'source', '$.message.value.to'),
			    jsonb_build_object('target', 'subject', 'source', '$.message.value.subject'),
			    jsonb_build_object('target', 'body', 'source', format('$.steps.%s.output.body', generator_step.id))
		    ),
		    updated_at = CURRENT_TIMESTAMP
		FROM generator_step
		WHERE ws.id = (SELECT id FROM summary_step);

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
			  AND version_number = 2
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		generator_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Генерация письма'
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		email_step AS (
			SELECT id FROM workflow_step
			WHERE workflow_version_id = (SELECT version_id FROM ver)
			  AND name = 'Отправка письма'
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		)
		UPDATE workflow_step ws
		SET input_mapping = jsonb_build_array(
			    jsonb_build_object('target', 'to', 'source', '$.message.value.to'),
			    jsonb_build_object('target', 'body', 'source', format('$.steps.%s.output.body', generator_step.id)),
			    jsonb_build_object('target', 'subject', 'source', '$.message.value.subject')
		    ),
		    updated_at = CURRENT_TIMESTAMP
		FROM generator_step
		WHERE ws.id = (SELECT id FROM email_step);

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		version_steps AS (
			SELECT ws.id
			FROM workflow_step ws
			JOIN workflow_version wv ON wv.id = ws.workflow_version_id
			WHERE wv.workflow_id = (SELECT id FROM wf)
			  AND wv.version_number IN (1, 2)
			  AND wv.deleted_at IS NULL
		)
		DELETE FROM workflow_step_dependency dep
		USING version_steps
		WHERE dep.step_id = version_steps.id
		   OR dep.depends_on_step_id = version_steps.id;

		WITH wf AS (
			SELECT id
			FROM workflow
			WHERE name = 'Уведомление о заявке'
			  AND system_id = (SELECT id FROM system WHERE name = 'Тестовая система' LIMIT 1)
			  AND deleted_at IS NULL
			ORDER BY id
			LIMIT 1
		),
		steps AS (
			SELECT wv.version_number, ws.id, ws.name
			FROM workflow_step ws
			JOIN workflow_version wv ON wv.id = ws.workflow_version_id
			WHERE wv.workflow_id = (SELECT id FROM wf)
			  AND wv.version_number IN (1, 2)
			  AND wv.deleted_at IS NULL
			  AND ws.deleted_at IS NULL
		),
		edges AS (
			SELECT *
			FROM (VALUES
				(1::int, 'Генерация письма'::text, 'Старт системы'::text, 'success'::text, 0::int),
				(1::int, 'Итоговое письмо'::text, 'Генерация письма'::text, 'success'::text, 0::int),
				(2::int, 'Генерация письма'::text, 'Старт системы'::text, 'success'::text, 0::int),
				(2::int, 'Проверка email'::text, 'Генерация письма'::text, 'success'::text, 0::int),
				(2::int, 'Письмо админа'::text, 'Проверка email'::text, 'true'::text, 0::int),
				(2::int, 'Отправка письма'::text, 'Проверка email'::text, 'false'::text, 0::int)
			) AS e(version_number, target_step_name, source_step_name, outcome, output_index)
		)
		INSERT INTO workflow_step_dependency (step_id, depends_on_step_id, outcome, output_index)
		SELECT target_step.id, source_step.id, edges.outcome, edges.output_index
		FROM edges
		JOIN steps target_step ON target_step.version_number = edges.version_number
			AND target_step.name = edges.target_step_name
		JOIN steps source_step ON source_step.version_number = edges.version_number
			AND source_step.name = edges.source_step_name
		ON CONFLICT (step_id, depends_on_step_id) DO UPDATE
		SET outcome = EXCLUDED.outcome,
		    output_index = EXCLUDED.output_index
	`)
	if err != nil {
		return fmt.Errorf("seed workflow steps: %w", err)
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
