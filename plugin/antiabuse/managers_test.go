package antiabuse

import "testing"

func TestManagers(t *testing.T) {
	err := managers.DoBlock(123)
	if err != nil {
		t.Fatal(err)
	}
	if !managers.IsBlocked(123) {
		t.Fatal("123 should be blocked but not")
	}
	err = managers.DoUnblock(123)
	if err != nil {
		t.Fatal(err)
	}
}
