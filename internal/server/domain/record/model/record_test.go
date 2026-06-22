package model

import "testing"

func TestRecordTypeConstants(t *testing.T) {
	t.Parallel()

	types := []RecordType{
		RecordTypeLogin, RecordTypeText, RecordTypeBinary, RecordTypeCard, RecordTypeOTP,
	}
	for _, typ := range types {
		if typ == "" {
			t.Fatal("empty record type")
		}
	}
}
