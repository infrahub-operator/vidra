package infrahub

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsValidTargetDateFormat", func() {
	It("accepts RFC3339 format", func() {
		isValid := IsValidTargetDateFormat("2024-01-01T15:04:05Z")
		Expect(isValid).To(BeTrue())
	})

	It("accepts relative format", func() {
		isValid := IsValidTargetDateFormat("now-2h")
		Expect(isValid).To(BeTrue())
	})

	It("rejects invalid format", func() {
		isValid := IsValidTargetDateFormat("yesterday")
		Expect(isValid).To(BeFalse())
	})
})

var _ = Describe("BuildURL", func() {
	It("replaces path params and appends query params", func() {
		url, err := BuildURL(
			"https://api.example.com",
			"/api/artifact/:id",
			map[string]string{"id": "123"},
			map[string]string{"branch": "main", "at": "now-1h"},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(url).To(ContainSubstring("/api/artifact/123"))
		Expect(url).To(ContainSubstring("branch=main"))
		Expect(url).To(ContainSubstring("at=now-1h"))
	})

	It("validates 'at' query param", func() {
		_, err := BuildURL(
			"https://api.example.com",
			"/api/artifact/:id",
			map[string]string{"id": "123"},
			map[string]string{"branch": "main", "at": "invalid"},
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid 'at' query param format"))
	})
	It("replaces path parameters and appends query params", func() {
		url, err := BuildURL("https://api.example.com", "/v1/resource/:id", map[string]string{
			"id": "42",
		}, map[string]string{
			"filter": "active",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://api.example.com/v1/resource/42?filter=active"))
	})

	It("encodes path and query parameters", func() {
		url, err := BuildURL("https://api.example.com", "/v1/resource/:id", map[string]string{
			"id": "complex id/with spaces",
		}, map[string]string{
			"q": "hello world",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://api.example.com/v1/resource/complex%20id%2Fwith%20spaces?q=hello+world"))
	})

	It("skips empty query parameters", func() {
		url, err := BuildURL("https://api.example.com", "/test", nil, map[string]string{
			"empty": "",
			"valid": "ok",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://api.example.com/test?valid=ok"))
	})

	It("validates 'at' parameter format", func() {
		url, err := BuildURL("https://api.example.com", "/at-test", nil, map[string]string{
			"at": "2025-04-09T00:00:00Z",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://api.example.com/at-test?at=2025-04-09T00%3A00%3A00Z"))
	})

	It("returns error on invalid 'at' format", func() {
		url, err := BuildURL("https://api.example.com", "/at-test", nil, map[string]string{
			"at": "not-a-date",
		})

		Expect(err).To(HaveOccurred())
		Expect(url).To(BeEmpty())
		Expect(err.Error()).To(ContainSubstring("invalid 'at' query param format"))
	})
})

var _ = Describe("infrahubClient", func() {
	var server *httptest.Server
	var client = NewClient()

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("Login", func() {
		It("returns token on successful login", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/api/auth/login"))
				Expect(r.Method).To(Equal("POST"))

				body := make(map[string]string)
				err := json.NewDecoder(r.Body).Decode(&body)
				Expect(err).ToNot(HaveOccurred())

				Expect(body["username"]).To(Equal("user"))
				Expect(body["password"]).To(Equal("pass"))

				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				err = json.NewEncoder(w).Encode(map[string]string{"access_token": "abc123"})
				Expect(err).ToNot(HaveOccurred())

			}))

			token, err := client.Login(server.URL, "user", "pass")
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("abc123"))
		})

		It("fails on bad credentials", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte(`invalid credentials`))
				Expect(err).ToNot(HaveOccurred())
			}))

			token, err := client.Login(server.URL, "user", "wrongpass")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("login failed with status"))
			Expect(token).To(BeEmpty())
		})
		It("returns error if response body is empty", func() {
			// Create a test server that returns 200 OK but empty body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Write no body (empty)
			}))
			defer server.Close()

			client := &infrahubClient{}

			token, err := client.Login(server.URL, "user", "pass")

			Expect(err).To(HaveOccurred())
			Expect(token).To(BeEmpty())
			Expect(err.Error()).To(ContainSubstring("EOF")) // JSON decode error usually returns EOF on empty body
		})
	})

	var _ = Describe("infrahubClient RunQuery", func() {
		var server *httptest.Server
		var client = NewClient()

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("succeeds with valid input", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("POST"))
				Expect(r.URL.Path).To(Equal("/api/query/test-query"))
				Expect(r.URL.Query().Get("branch")).To(Equal("main"))
				Expect(r.URL.Query().Get("at")).To(Equal("2025-01-01T00:00:00Z"))

				var payload map[string]interface{}
				Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
				Expect(payload["variables"]).To(HaveKeyWithValue("artifactname", "test-artifact"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"CoreArtifact": map[string]interface{}{
							"edges": []map[string]interface{}{
								{
									"node": map[string]interface{}{
										"id": "a1",
										"storage_id": map[string]interface{}{
											"id": "s1",
										},
										"checksum": map[string]interface{}{
											"value": "abc123",
										},
									},
								},
							},
						},
					},
				})
			}))

			client := &infrahubClient{}

			result, err := client.RunQuery("test-query", server.URL, "test-artifact", "main", "2025-01-01T00:00:00Z", "token123")
			Expect(err).ToNot(HaveOccurred())
			Expect(*result).To(HaveLen(1))
			Expect((*result)[0].ID).To(Equal("a1"))
			Expect((*result)[0].StorageID).To(Equal("s1"))
			Expect((*result)[0].Checksum).To(Equal("abc123"))
		})

		It("fails on BuildURL error", func() {
			result, err := client.RunQuery("test-query", "://invalid-url", "a", "b", "notadate", "token")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to build query URL"))
		})

		It("fails on non-200 response", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte(`{"error": "bad request"}`))
				Expect(err).ToNot(HaveOccurred())
			}))

			result, err := client.RunQuery("test-query", server.URL, "a", "b", "2025-01-01T00:00:00Z", "token")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("query failed with status"))
		})

		It("fails on bad JSON response", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{bad json}`))
				Expect(err).ToNot(HaveOccurred())
			}))

			result, err := client.RunQuery("test-query", server.URL, "a", "b", "2025-01-01T00:00:00Z", "token")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to decode query result"))
		})
		var _ = Describe("infrahubClient RunQuery", func() {
			It("returns error when response body is empty", func() {
				// Setup test server returning 200 OK but empty body
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// No body content
				}))
				defer server.Close()

				client := &infrahubClient{}

				_, err := client.RunQuery("test-query", server.URL, "test-artifact", "main", "2025-01-01T00:00:00Z", "token123")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("EOF")) // Likely JSON decoding will fail due to empty body
			})
		})

	})
	var _ = Describe("infrahubClient.DownloadArtifact", func() {
		var (
			server     *httptest.Server
			client     *infrahubClient
			apiURL     string
			artifactID string
			branch     string
			date       string
		)

		BeforeEach(func() {
			artifactID = "test-artifact-id"
			branch = "main"
			date = "2025-01-01T00:00:00Z"
			client = &infrahubClient{}
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("downloads the artifact successfully", func() {
			// Create mock server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal("/api/artifact/test-artifact-id"))
				Expect(r.URL.Query().Get("branch")).To(Equal("main"))
				Expect(r.URL.Query().Get("at")).To(Equal("2025-01-01T00:00:00Z"))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("artifact-content"))
			}))

			apiURL = server.URL
			reader, err := client.DownloadArtifact(apiURL, artifactID, branch, date, "mock-token")
			Expect(err).NotTo(HaveOccurred())

			content, err := io.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("artifact-content"))
		})

		It("fails to send GET request with malformed URL", func() {
			apiURL = ":::invalid-url"
			reader, err := client.DownloadArtifact(apiURL, artifactID, branch, date, "mock-token")
			Expect(reader).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create request"))

		})

		It("fails to send GET request", func() {
			// Use a non-routable address to trigger http.Get error
			apiURL = "http://127.0.0.1:0" // closed port
			reader, err := client.DownloadArtifact(apiURL, artifactID, branch, date, "mock-token")
			Expect(reader).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to download artifact after retries"))

		})

		It("returns error on non-200 response", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			apiURL = server.URL

			content, err := client.DownloadArtifact(apiURL, artifactID, branch, date, "token123")
			Expect(err).To(HaveOccurred())
			Expect(content).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to download artifact"))
		})

		It("returns error on malformed URL", func() {
			content, err := client.DownloadArtifact(":://badurl", artifactID, branch, date, "token")
			Expect(err).To(HaveOccurred())
			Expect(content).To(BeNil())
		})
		It("fails to build the artifact URL", func() {
			apiURL = "://bad-url"
			body, err := client.DownloadArtifact(apiURL, artifactID, branch, date, "token123")
			Expect(err).To(HaveOccurred())
			Expect(body).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to create request"))
		})

		It("fails when the server returns non-200", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not found", http.StatusNotFound)
			}))

			apiURL = server.URL
			body, err := client.DownloadArtifact(apiURL, artifactID, branch, date, "token123")
			Expect(err).To(HaveOccurred())
			Expect(body).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("last status code: 404"))
		})
		It("returns an error if the date format is invalid", func() {

			_, err := client.DownloadArtifact(
				"http://example.com",
				"artifact123",
				"main",
				"invalid-date-format", // intentionally wrong
				"token123",
			)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to build URL for artifact"))
			// Optionally check for date format specific error if BuildURL returns that
		})

	})

})
