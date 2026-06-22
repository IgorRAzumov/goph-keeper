package model

import "testing"

func TestNewIDIsUniqueAndNonEmpty(t *testing.T) {
	t.Parallel()

	first, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	if first == "" || second == "" {
		t.Fatal("expected non-empty id")
	}
	if first == second {
		t.Fatalf("expected unique ids, got %q twice", first)
	}
}

func TestRecordTypeValues(t *testing.T) {
	t.Parallel()

	cases := map[RecordType]string{
		RecordTypeLogin:  "login",
		RecordTypeText:   "text",
		RecordTypeBinary: "binary",
		RecordTypeCard:   "card",
		RecordTypeOTP:    "otp",
	}
	for rt, want := range cases {
		if string(rt) != want {
			t.Errorf("RecordType %q != %q", rt, want)
		}
	}
}
