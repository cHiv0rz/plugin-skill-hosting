package email

import (
	"strings"
	"testing"
)

func TestConfigured(t *testing.T) {
	if (Sender{}).Configured() {
		t.Error("zero Sender should not be configured")
	}
	if (Sender{Host: "smtp.example.com"}).Configured() {
		t.Error("missing From should not be configured")
	}
	if !(Sender{Host: "smtp.example.com", From: "a@b.com"}).Configured() {
		t.Error("host+from should be configured")
	}
}

func TestBuildMessage(t *testing.T) {
	msg := string(buildMessage("from@x.com", []string{"a@x.com", "b@x.com"}, "Hello", "line1\nline2"))
	for _, want := range []string{
		"From: from@x.com\r\n",
		"To: a@x.com, b@x.com\r\n",
		"Subject: Hello\r\n",
		"Content-Type: text/plain; charset=UTF-8\r\n",
		"\r\nline1\r\nline2\r\n",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q\n--- full ---\n%s", want, msg)
		}
	}
}

func TestBuildMessage_DotStuffing(t *testing.T) {
	msg := string(buildMessage("f@x.com", []string{"t@x.com"}, "s", ".hidden"))
	if !strings.Contains(msg, "\r\n..hidden\r\n") {
		t.Errorf("leading dot not stuffed: %q", msg)
	}
}
