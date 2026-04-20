package calculatorsession

import "testing"

func TestNormalizeFields(t *testing.T) {
	got := normalizeFields([]string{"1", "2"})
	if len(got) != fieldCount {
		t.Fatalf("len %d", len(got))
	}
	if got[0] != "1" || got[1] != "2" {
		t.Fatal(got)
	}
	for i := 2; i < fieldCount; i++ {
		if got[i] != "" {
			t.Fatalf("index %d", i)
		}
	}
}

func TestMigrateNineToTen(t *testing.T) {
	old := []string{"5", "1", "3", "1", "90", "1000000", "3000", "500", "1"}
	got := migrateCalculatorFields(old)
	if got == nil || len(got) != fieldCount {
		t.Fatalf("migrate: %v", got)
	}
	if got[0] != "4" || got[1] != "1" {
		t.Fatalf("data nodes / masters: %v %v", got[0], got[1])
	}
}

func TestMigrateTwelveToTen(t *testing.T) {
	old := []string{"2", "0", "1", "0", "10", "1", "50", "8", "1", "0.95", "32", "500"}
	got := migrateCalculatorFields(old)
	if got == nil || len(got) != fieldCount {
		t.Fatalf("migrate: %v", got)
	}
	if got[8] != "32" || got[9] != "500" {
		t.Fatalf("ram/disk: %v %v", got[8], got[9])
	}
}
