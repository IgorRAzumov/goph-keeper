package record

import "goph-keeper/internal/domain/record/model"

type GetRecordInput struct {
	OwnerID  string
	RecordID string
}

type GetRecordOutput struct {
	Record *model.Record
}

type UpdateRecordInput struct {
	OwnerID string
	Record  *model.Record
}

type UpdateRecordOutput struct{}
