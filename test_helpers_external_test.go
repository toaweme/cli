package cli_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("%swant %v, got %v", formatMsg(msgAndArgs), expected, actual)
	}
}

func assertNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		t.Fatalf("%sunexpected error: %v", formatMsg(msgAndArgs), err)
	}
}

func assertErrorIs(t *testing.T, err, target error, msgAndArgs ...any) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("%serror %v does not match %v", formatMsg(msgAndArgs), err, target)
	}
}

func assertNotNil(t *testing.T, val any, msgAndArgs ...any) {
	t.Helper()
	if val == nil || reflect.ValueOf(val).IsNil() {
		t.Fatalf("%sexpected non-nil", formatMsg(msgAndArgs))
	}
}

func assertTrue(t *testing.T, val bool, msgAndArgs ...any) {
	t.Helper()
	if !val {
		t.Fatalf("%sexpected true", formatMsg(msgAndArgs))
	}
}

func assertContains(t *testing.T, s, substr string, msgAndArgs ...any) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("%sexpected output to contain %q", formatMsg(msgAndArgs), substr)
	}
}

func assertNotContains(t *testing.T, s, substr string, msgAndArgs ...any) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("%sexpected output to not contain %q", formatMsg(msgAndArgs), substr)
	}
}

func assertLen(t *testing.T, val any, expected int, msgAndArgs ...any) {
	t.Helper()
	v := reflect.ValueOf(val)
	if v.Len() != expected {
		t.Fatalf("%sexpected length %d, got %d", formatMsg(msgAndArgs), expected, v.Len())
	}
}

func assertGreater(t *testing.T, a, b int, msgAndArgs ...any) {
	t.Helper()
	if a <= b {
		t.Fatalf("%sexpected %d > %d", formatMsg(msgAndArgs), a, b)
	}
}

func assertValidJSON(t *testing.T, s string, msgAndArgs ...any) {
	t.Helper()
	var v []json.RawMessage
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("%sinvalid JSON: %v", formatMsg(msgAndArgs), err)
	}
}

func unmarshalJSON(t *testing.T, s string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(s), v); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\n%s", err, s)
	}
}

func formatMsg(msgAndArgs []any) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return fmt.Sprintf("%v: ", msgAndArgs[0])
	}
	return fmt.Sprintf(msgAndArgs[0].(string)+": ", msgAndArgs[1:]...)
}
