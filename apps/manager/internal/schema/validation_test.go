package schema

import "testing"

func TestValidateRawAcceptsPayloadMatchingDialectRequired(t *testing.T) {
	err := ValidateRaw(
		[]byte(`{"type":"object","properties":{"host":{"type":"string","required":true},"port":{"type":"integer","required":true}}}`),
		[]byte(`{"host":"smtp.local","port":1025}`),
	)
	if err != nil {
		t.Fatalf("ValidateRaw error: %v", err)
	}
}

func TestValidateRawRejectsMissingRequiredField(t *testing.T) {
	err := ValidateRaw(
		[]byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
		[]byte(`{}`),
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRawRejectsWrongType(t *testing.T) {
	err := ValidateRaw(
		[]byte(`{"type":"object","properties":{"port":{"type":"integer","required":true}}}`),
		[]byte(`{"port":"1025"}`),
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRawRejectsInvalidSchema(t *testing.T) {
	err := ValidateRaw(
		[]byte(`{"type":`),
		[]byte(`{"port":1025}`),
	)
	if err == nil {
		t.Fatal("expected invalid schema error")
	}
}
