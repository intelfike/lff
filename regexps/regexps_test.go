package regexps

import "testing"

func TestReplaceAllRegexp(t *testing.T) {
	got, matched := ReplaceAll("func main() {}", []string{`func\s+main`}, "func entry", false)
	if !matched {
		t.Fatal("expected regexp replacement to match")
	}
	if want := "func entry() {}"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReplaceAllFixed(t *testing.T) {
	got, matched := ReplaceAll("hello, world", []string{"world"}, "gopher", true)
	if !matched {
		t.Fatal("expected plain text replacement to match")
	}
	if want := "hello, gopher"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReplaceAllNoMatch(t *testing.T) {
	got, matched := ReplaceAll("hello, world", []string{"nope"}, "gopher", true)
	if matched {
		t.Fatal("expected no match")
	}
	if want := "hello, world"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReplaceAllMultiplePatterns(t *testing.T) {
	got, matched := ReplaceAll("foo bar baz", []string{"foo", "baz"}, "X", true)
	if !matched {
		t.Fatal("expected match")
	}
	if want := "X bar X"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
