package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v93/github"
	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/go-retryablehttp"
	v3 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v3"
	v4 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v4"
)

type Client struct {
	v4Client V4Client
	v3Client V3Client
}

type V4Client interface {
	GetPR(ctx context.Context, owner, name string, number int) (*v4.PullRequest, error)
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
}

type V3Client interface {
	CompareCommits(ctx context.Context, owner, repo, base, head string) ([]string, error)
	IsAncestor(ctx context.Context, owner, repo, ancestor, descendant string) (bool, error)
}

type (
	PullRequestReviewEvent = github.PullRequestReviewEvent
	PullRequestEvent       = github.PullRequestEvent
	CheckSuiteEvent        = github.CheckSuiteEvent
)

// requestTimeout is the timeout of each attempt of an HTTP request to GitHub API.
// Each retry has its own timeout.
// It prevents a stalled connection from blocking webhook processing forever.
// GitHub terminates API requests that take more than 10 seconds to process,
// so waiting much longer than that is pointless.
// It is also short enough that a retry can start within the HTTP server's graceful shutdown (25 seconds),
// and finish in time if it succeeds quickly. It doesn't bound the total time of a request with retries.
// https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api
// https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api
// Note that it doesn't bound the wait between retries:
// go-retryablehttp waits as long as the Retry-After header says on 429 and 503 responses.
const requestTimeout = 15 * time.Second

// ParamNewApp is the parameter of New.
// The GitHub API clients authenticate as the installation of the GitHub App.
type ParamNewApp struct {
	AppID int64
	// Signer signs JWTs to get installation access tokens.
	Signer         Signer
	InstallationID int64
	Logger         *slog.Logger
}

// Signer signs JWTs for the GitHub App.
type Signer = ghinstallation.Signer

// NewRSASigner creates a Signer from the GitHub App's private key in PEM format.
func NewRSASigner(privateKey string) (Signer, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("parse the GitHub App's private key: %w", err)
	}
	return ghinstallation.NewRSASigner(jwt.SigningMethodRS256, key), nil
}

var ValidateSignature = github.ValidateSignature //nolint:gochecknoglobals

// New creates a client of GitHub REST and GraphQL APIs.
// Both share one HTTP client, so they share the authentication, retry, and timeout settings.
func New(param *ParamNewApp) (*Client, error) {
	httpClient, err := newHTTPClient(param)
	if err != nil {
		return nil, fmt.Errorf("create an HTTP client for GitHub API: %w", err)
	}
	v3Client, err := v3.New(httpClient)
	if err != nil {
		return nil, fmt.Errorf("create GitHub v3 client: %w", err)
	}
	return &Client{
		v4Client: v4.New(httpClient),
		v3Client: v3Client,
	}, nil
}

// newHTTPClient creates an HTTP client shared by the GitHub REST and GraphQL API clients.
// It authenticates as the GitHub App installation, retries failed requests, and times out each attempt.
func newHTTPClient(param *ParamNewApp) (*http.Client, error) {
	atr, err := ghinstallation.NewAppsTransportWithOptions(http.DefaultTransport, param.AppID, ghinstallation.WithSigner(param.Signer))
	if err != nil {
		return nil, fmt.Errorf("create a transport for the GitHub App: %w", err)
	}
	itr := ghinstallation.NewFromAppsTransport(atr, param.InstallationID)
	c := retryablehttp.NewClient()
	c.HTTPClient = &http.Client{Transport: itr, Timeout: requestTimeout}
	c.Logger = param.Logger
	c.CheckRetry = checkRetry
	return c.StandardClient(), nil
}

// checkRetry is the retry policy of HTTP requests to GitHub API.
// It doesn't retry if signing a JWT fails, otherwise it follows the default retry policy.
// A signing error is permanent in most cases, such as a missing kms:Sign permission or a wrong key ID,
// and transient errors of AWS KMS are already retried by the AWS SDK.
// Retrying it only calls kms:Sign repeatedly and delays the failure.
func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if isSignError(err) {
		return false, err
	}
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err) //nolint:wrapcheck
}

// signErrorPrefix is the prefix of the error message when ghinstallation fails to sign a JWT.
// ghinstallation formats the error from Signer with %s, so the error can't be detected with errors.As.
// https://github.com/bradleyfalzon/ghinstallation/blob/v2.19.0/appsTransport.go#L98-L101
const signErrorPrefix = "could not sign jwt: "

// isSignError returns true if err is caused by a failure to sign a JWT.
// ghinstallation returns HTTPError without a response when it fails to get an installation access token.
func isSignError(err error) bool {
	var httpErr *ghinstallation.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	return httpErr.Response == nil && httpErr.RootCause != nil && strings.HasPrefix(httpErr.RootCause.Error(), signErrorPrefix)
}
