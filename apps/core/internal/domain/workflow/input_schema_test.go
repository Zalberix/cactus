package workflow

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

func TestUpsertWorkflowInputSchemaFieldUsesPropertyRequiredDialect(t *testing.T) {
	schema, err := upsertWorkflowInputSchemaField(nil, InputSchemaFieldRequest{
		Name:        "fio",
		Type:        "string",
		Required:    true,
		Description: "Full name",
	})
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"type":"object",
		"properties":{
			"fio":{"type":"string","required":true,"description":"Full name"}
		}
	}`, string(schema))
}

func TestUpsertWorkflowInputSchemaFieldRejectsNestedName(t *testing.T) {
	_, err := upsertWorkflowInputSchemaField(nil, InputSchemaFieldRequest{
		Name: "customer.name",
		Type: "string",
	})
	require.Error(t, err)
}

func TestDeleteWorkflowInputSchemaFieldRemovesOnlyRequestedField(t *testing.T) {
	schema, err := deleteWorkflowInputSchemaField([]byte(`{
		"type":"object",
		"properties":{
			"email":{"type":"string","required":true},
			"subject":{"type":"string"}
		}
	}`), "email")
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"type":"object",
		"properties":{
			"subject":{"type":"string"}
		}
	}`, string(schema))
}

func TestDeleteWorkflowInputSchemaFieldRejectsMissingField(t *testing.T) {
	_, err := deleteWorkflowInputSchemaField([]byte(`{"type":"object","properties":{}}`), "email")
	require.Error(t, err)
}

func TestParseTopLevelPropertiesUsesPropertyRequiredDialect(t *testing.T) {
	props, err := parseTopLevelProperties([]byte(`{
		"type":"object",
		"required":["ignored"],
		"properties":{
			"to":{"type":"string","required":true},
			"cc":{"type":"string"}
		}
	}`))
	require.NoError(t, err)

	assert.Equal(t, schemaProperty{Type: "string", Required: true}, props["to"])
	assert.Equal(t, schemaProperty{Type: "string", Required: false}, props["cc"])
}

func TestParseTopLevelPropertiesRejectsNestedName(t *testing.T) {
	_, err := parseTopLevelProperties([]byte(`{
		"type":"object",
		"properties":{"customer.name":{"type":"string"}}
	}`))
	require.Error(t, err)
}

func TestValidateStepInputsFilled(t *testing.T) {
	workflowSchema := []byte(`{
		"type":"object",
		"properties":{
			"to":{"type":"string","required":true},
			"optional_to":{"type":"string"}
		}
	}`)

	tests := []struct {
		name      string
		steps     []db.ListEnrichedStepsByVersionIDRow
		deps      []db.WorkflowStepDependency
		wantValid bool
	}{
		{
			name: "required target from message value with required workflow field valid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, `[{"target":"to","source":"$.message.value.to"}]`),
			},
			wantValid: true,
		},
		{
			name: "missing workflow field invalid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, `[{"target":"to","source":"$.message.value.missing"}]`),
			},
		},
		{
			name: "optional workflow field invalid for required target",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, `[{"target":"to","source":"$.message.value.optional_to"}]`),
			},
		},
		{
			name: "required target from required source output valid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, nil, `{"type":"object","properties":{"body":{"type":"string","required":true}}}`, nil),
				taskRow(2, `{"type":"object","properties":{"body":{"type":"string","required":true}}}`, nil, `[{"target":"body","source":"$.steps.1.output.body"}]`),
			},
			deps:      []db.WorkflowStepDependency{depRow(2, 1)},
			wantValid: true,
		},
		{
			name: "optional source output invalid for required target",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, nil, `{"type":"object","properties":{"body":{"type":"string"}}}`, nil),
				taskRow(2, `{"type":"object","properties":{"body":{"type":"string","required":true}}}`, nil, `[{"target":"body","source":"$.steps.1.output.body"}]`),
			},
			deps: []db.WorkflowStepDependency{depRow(2, 1)},
		},
		{
			name: "static source valid for required string",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"subject":{"type":"string","required":true}}}`, nil, `[{"target":"subject","source":"Welcome"}]`),
			},
			wantValid: true,
		},
		{
			name: "required target without mapping invalid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, nil),
			},
		},
		{
			name: "nested message source invalid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, `[{"target":"to","source":"$.message.value.customer.name"}]`),
			},
		},
		{
			name: "nested target invalid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, `[{"target":"customer.name","source":"$.message.value.to"}]`),
			},
		},
		{
			name: "duplicate target invalid",
			steps: []db.ListEnrichedStepsByVersionIDRow{
				taskRow(1, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, nil, `[{"target":"to","source":"$.message.value.to"},{"target":"to","source":"Welcome"}]`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateStepInputsFilled(tt.steps, tt.deps, workflowSchema)
			if tt.wantValid {
				assert.Empty(t, errs)
			} else {
				require.NotEmpty(t, errs)
				assert.Equal(t, "invalid_mapping", errs[0].Type)
			}
		})
	}
}

func taskRow(id int32, inputSchema, outputSchema, mapping any) db.ListEnrichedStepsByVersionIDRow {
	return db.ListEnrichedStepsByVersionIDRow{
		ID:                id,
		WorkflowVersionID: 10,
		StepType:          "task",
		InputSchema:       rawBytes(inputSchema),
		OutputSchema:      rawBytes(outputSchema),
		InputMapping:      rawBytes(mapping),
	}
}

func depRow(stepID, dependsOnStepID int32) db.WorkflowStepDependency {
	return db.WorkflowStepDependency{
		StepID:          stepID,
		DependsOnStepID: dependsOnStepID,
		Outcome:         pgtype.Text{String: "success", Valid: true},
	}
}

func rawBytes(v any) []byte {
	switch value := v.(type) {
	case nil:
		return nil
	case string:
		return []byte(value)
	default:
		return nil
	}
}
