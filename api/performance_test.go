package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/cli/cli/v2/pkg/httpmock"
)

// BenchmarkGraphQLQueries benchmarks GraphQL query performance
func BenchmarkGraphQLQueries(b *testing.B) {
	http := &httpmock.Registry{}
	client := newTestClient(http)

	vars := map[string]interface{}{"owner": "cli", "repo": "cli"}
	response := struct {
		Repository struct {
			Name string
		}
	}{}

	http.Register(
		httpmock.GraphQL("QUERY"),
		httpmock.StringResponse(`{"data":{"repository":{"name":"cli"}}}`),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GraphQL("github.com", "QUERY", vars, &response)
	}
}

// BenchmarkRESTRequests benchmarks REST API request performance
func BenchmarkRESTRequests(b *testing.B) {
	http := &httpmock.Registry{}
	client := newTestClient(http)

	http.Register(
		httpmock.REST("GET", "repos/cli/cli"),
		httpmock.StringResponse(`{"name":"cli","full_name":"cli/cli"}`),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.REST("github.com", "GET", "repos/cli/cli", nil, nil)
	}
}

// BenchmarkCachedHTTPClient benchmarks cached HTTP client performance
func BenchmarkCachedHTTPClient(b *testing.B) {
	baseClient := &http.Client{}
	ttl := 5 * time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCachedHTTPClient(baseClient, ttl)
	}
}

// BenchmarkAddAuthTokenHeader benchmarks auth token header addition
func BenchmarkAddAuthTokenHeader(b *testing.B) {
	rt := http.DefaultTransport
	cfg := &mockTokenGetter{token: "gho_test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AddAuthTokenHeader(rt, cfg)
	}
}

type mockTokenGetter struct {
	token string
}

func (m *mockTokenGetter) ActiveToken(host string) (string, string) {
	return m.token, "oauth"
}
