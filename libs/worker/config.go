package worker

import (
	"context"
	"encoding/json"
	"fmt"
)

type configPayload struct {
	OrganizationID int32          `json:"organization_id"`
	WorkTypeID     int32          `json:"work_type_id"`
	SchemaID       int32          `json:"schema_id"`
	RevisionID     int32          `json:"revision_id"`
	ConfigHash     string         `json:"config_hash"`
	SettingsData   map[string]any `json:"settings_data"`
}

func (w *Worker) resolveSettings(ctx context.Context, ref ConfigRef) (map[string]any, error) {
	if ref.OrganizationID != w.cfg.OrganizationID || ref.WorkTypeID != w.cfg.WorkTypeID {
		return nil, fmt.Errorf("config_ref scope mismatch")
	}
	if settings, ok := w.configCache.get(ref); ok {
		return settings, nil
	}
	payload, err := w.readConfigFromNATS(ctx, ref)
	if err != nil {
		return nil, err
	}
	w.configCache.set(ref, payload.SettingsData)
	return payload.SettingsData, nil
}

func (w *Worker) readConfigFromNATS(ctx context.Context, ref ConfigRef) (configPayload, error) {
	payload, err := w.readConfigFromStream(ctx, ref)
	if err == nil {
		return payload, nil
	}
	payload, err = w.requestConfig(ctx, ref)
	if err != nil {
		return configPayload{}, err
	}
	return payload, nil
}

func (w *Worker) readConfigFromStream(ctx context.Context, ref ConfigRef) (configPayload, error) {
	stream, err := w.js.Stream(ctx, "CONFIGS")
	if err != nil {
		return configPayload{}, fmt.Errorf("open CONFIGS stream: %w", err)
	}
	raw, err := stream.GetLastMsgForSubject(ctx, configSubject(ref))
	if err != nil {
		return configPayload{}, fmt.Errorf("read config subject: %w", err)
	}
	return decodeConfigPayload(raw.Data, ref)
}

func (w *Worker) requestConfig(ctx context.Context, ref ConfigRef) (configPayload, error) {
	msg, err := w.nc.RequestWithContext(ctx, configRequestSubject(ref), nil)
	if err != nil {
		return configPayload{}, fmt.Errorf("request config: %w", err)
	}
	return decodeConfigPayload(msg.Data, ref)
}

func decodeConfigPayload(data []byte, ref ConfigRef) (configPayload, error) {
	var payload configPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return configPayload{}, fmt.Errorf("unmarshal config payload: %w", err)
	}
	if payload.OrganizationID != ref.OrganizationID ||
		payload.WorkTypeID != ref.WorkTypeID ||
		payload.SchemaID != ref.SchemaID ||
		payload.RevisionID != ref.RevisionID {
		return configPayload{}, fmt.Errorf("config payload scope mismatch")
	}
	if payload.ConfigHash != ref.ConfigHash {
		return configPayload{}, fmt.Errorf("config payload hash mismatch")
	}
	return payload, nil
}

func configSubject(ref ConfigRef) string {
	if ref.ConfigSubject != "" {
		return ref.ConfigSubject
	}
	return fmt.Sprintf("config.org.%d.work_type.%d.revision.%d", ref.OrganizationID, ref.WorkTypeID, ref.RevisionID)
}

func configRequestSubject(ref ConfigRef) string {
	return fmt.Sprintf("config.request.org.%d.work_type.%d.revision.%d", ref.OrganizationID, ref.WorkTypeID, ref.RevisionID)
}
