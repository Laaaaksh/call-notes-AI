package fieldmapper

import (
	"context"

	exEntities "github.com/call-notes-ai-service/internal/modules/extraction/entities"
	fmEntities "github.com/call-notes-ai-service/internal/modules/fieldmapper/entities"
)

type ICore interface {
	MapEntitiesToFields(ctx context.Context, sessionID string, extracted []exEntities.MedicalEntity) (*fmEntities.MappedFields, error)
	GetFieldMappings() []fmEntities.SalesforceFieldMapping
}

type Core struct {
	mappings []fmEntities.SalesforceFieldMapping
}

var _ ICore = (*Core)(nil)

func NewCore(_ context.Context) ICore {
	return &Core{
		mappings: defaultFieldMappings(),
	}
}

func (c *Core) MapEntitiesToFields(_ context.Context, sessionID string, extracted []exEntities.MedicalEntity) (*fmEntities.MappedFields, error) {
	fields := make(map[string]string)

	entityToField := buildEntityFieldMap()

	for _, entity := range extracted {
		if entity.IsNegated {
			continue
		}
		if sfField, ok := entityToField[string(entity.Type)]; ok {
			fields[sfField] = entity.NormalizedValue
		}
	}

	return &fmEntities.MappedFields{
		SessionID: sessionID,
		Fields:    fields,
	}, nil
}

func (c *Core) GetFieldMappings() []fmEntities.SalesforceFieldMapping {
	return c.mappings
}

func buildEntityFieldMap() map[string]string {
	return map[string]string{
		"name":       "Patient_Name__c",
		"age":        "Patient_Age__c",
		"phone":      "Patient_Phone__c",
		"gender":     "Patient_Gender__c",
		"symptom":    "Primary_Symptom__c",
		"body_part":  "Body_Part_Affected__c",
		"condition":  "Medical_Condition__c",
		"medication": "Current_Medication__c",
		"duration":   "Symptom_Duration__c",
		"severity":   "Severity_Level__c",
		"allergy":    "Known_Allergies__c",
		"follow_up":  "Follow_Up_Required__c",
		"icd10_code": "ICD10_Code__c",
	}
}

func defaultFieldMappings() []fmEntities.SalesforceFieldMapping {
	return []fmEntities.SalesforceFieldMapping{
		{FieldName: "patient_name", SFFieldAPI: "Patient_Name__c", DataType: "string", Required: true, MaxLength: 255},
		{FieldName: "patient_age", SFFieldAPI: "Patient_Age__c", DataType: "integer", Required: false},
		{FieldName: "patient_phone", SFFieldAPI: "Patient_Phone__c", DataType: "phone", Required: true, MaxLength: 15},
		{FieldName: "patient_gender", SFFieldAPI: "Patient_Gender__c", DataType: "picklist", Required: false},
		{FieldName: "primary_symptom", SFFieldAPI: "Primary_Symptom__c", DataType: "string", Required: true, MaxLength: 500},
		{FieldName: "body_part", SFFieldAPI: "Body_Part_Affected__c", DataType: "string", Required: false, MaxLength: 255},
		{FieldName: "condition", SFFieldAPI: "Medical_Condition__c", DataType: "string", Required: false, MaxLength: 500},
		{FieldName: "medication", SFFieldAPI: "Current_Medication__c", DataType: "string", Required: false, MaxLength: 500},
		{FieldName: "duration", SFFieldAPI: "Symptom_Duration__c", DataType: "string", Required: false, MaxLength: 100},
		{FieldName: "severity", SFFieldAPI: "Severity_Level__c", DataType: "picklist", Required: false},
		{FieldName: "allergies", SFFieldAPI: "Known_Allergies__c", DataType: "text", Required: false},
		{FieldName: "follow_up", SFFieldAPI: "Follow_Up_Required__c", DataType: "boolean", Required: false},
		{FieldName: "icd10_code", SFFieldAPI: "ICD10_Code__c", DataType: "string", Required: false, MaxLength: 10},
		{FieldName: "call_summary", SFFieldAPI: "Call_Summary__c", DataType: "textarea", Required: false},
		{FieldName: "agent_notes", SFFieldAPI: "Agent_Notes__c", DataType: "textarea", Required: false},
	}
}
