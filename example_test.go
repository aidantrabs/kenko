package kenko_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/aidantrabs/kenko"
)

func ExampleNew() {
	k, err := kenko.New(
		kenko.WithTarget("google", "https://google.com"),
		kenko.WithInterval(10 * time.Second),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(k.Checker().Ready())
	// Output: false
}

func ExampleKenko_RegisterHandlers() {
	k, err := kenko.New(
		kenko.WithTarget("example", "https://example.com"),
	)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	k.RegisterHandlers(mux)
	// mux now serves /health, /ready, and /status
}

func ExampleNewChecker() {
	c, err := kenko.NewChecker(
		kenko.WithTarget("api", "https://api.example.com/health"),
		kenko.WithInterval(15 * time.Second),
		kenko.WithTimeout(3 * time.Second),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(c.Ready())
	// Output: false
}

func ExampleNewMemoryStore() {
	store := kenko.NewMemoryStore()
	_ = store.Set(context.Background(), "api", kenko.Result{
		Target: "api",
		Status: kenko.StatusHealthy,
	})

	results, _ := store.GetAll(context.Background())
	fmt.Println(results["api"].Status)
	// Output: healthy
}

func ExampleHandleHealth() {
	checker, _ := kenko.NewChecker(
		kenko.WithTarget("example", "https://example.com"),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	kenko.HandleHealth(checker)(w, r)

	fmt.Println(w.Code)
	fmt.Print(w.Body.String())
	// Output:
	// 200
	// {"status":"healthy"}
}

func ExampleHandleReady() {
	checker, _ := kenko.NewChecker(
		kenko.WithTarget("example", "https://example.com"),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	kenko.HandleReady(checker)(w, r)

	fmt.Println(w.Code)
	fmt.Print(w.Body.String())
	// Output:
	// 503
	// {"status":"not_ready"}
}

func ExampleHandleStatus() {
	checker, _ := kenko.NewChecker(
		kenko.WithTarget("example", "https://example.com"),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/status", nil)
	kenko.HandleStatus(checker)(w, r)

	fmt.Println(w.Code)
	fmt.Print(w.Body.String())
	// Output:
	// 200
	// {"targets":[]}
}
