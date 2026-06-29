package cli

import (
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

func assertError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		t.Fatalf("%sexpected error, got nil", formatMsg(msgAndArgs))
	}
}

func assertErrorIs(t *testing.T, err, target error, msgAndArgs ...any) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("%serror %v does not match %v", formatMsg(msgAndArgs), err, target)
	}
}

func assertNil(t *testing.T, val any, msgAndArgs ...any) {
	t.Helper()
	if val == nil {
		return
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Chan || v.Kind() == reflect.Func {
		if v.IsNil() {
			return
		}
	}
	t.Fatalf("%sexpected nil, got %v", formatMsg(msgAndArgs), val)
}

func assertNotNil(t *testing.T, val any, msgAndArgs ...any) {
	t.Helper()
	if val == nil {
		t.Fatalf("%sexpected non-nil", formatMsg(msgAndArgs))
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Chan || v.Kind() == reflect.Func {
		if v.IsNil() {
			t.Fatalf("%sexpected non-nil", formatMsg(msgAndArgs))
		}
	}
}

func assertContains(t *testing.T, s, substr string, msgAndArgs ...any) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("%sexpected %q to contain %q", formatMsg(msgAndArgs), truncate(s, 200), substr)
	}
}

func assertEmpty(t *testing.T, val any, msgAndArgs ...any) {
	t.Helper()
	v := reflect.ValueOf(val)
	if v.Len() != 0 {
		t.Fatalf("%sexpected empty, got length %d", formatMsg(msgAndArgs), v.Len())
	}
}

func assertLen(t *testing.T, val any, expected int, msgAndArgs ...any) {
	t.Helper()
	v := reflect.ValueOf(val)
	if v.Len() != expected {
		t.Fatalf("%sexpected length %d, got %d", formatMsg(msgAndArgs), expected, v.Len())
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
