package redisstore

import (
	"testing"
)

func TestNew_DefaultKeyPrefix(t *testing.T) {
	s := New("localhost:6379")
	if s.keyPrefix != defaultKeyPrefix {
		t.Errorf("keyPrefix = %q, want %q", s.keyPrefix, defaultKeyPrefix)
	}
}

func TestNew_WithPassword(t *testing.T) {
	s := New("localhost:6379", WithPassword("secret"))
	if s.password != "secret" {
		t.Errorf("password = %q, want %q", s.password, "secret")
	}
}

func TestNew_WithKeyPrefix(t *testing.T) {
	s := New("localhost:6379", WithKeyPrefix("myapp:health"))
	if s.keyPrefix != "myapp:health" {
		t.Errorf("keyPrefix = %q, want %q", s.keyPrefix, "myapp:health")
	}
}
