package aws

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/golang-jwt/jwt/v5"
)

// kmsSignTimeout is the timeout of a kms:Sign call, including the retries by the AWS SDK.
// Signing runs inside an HTTP request to GitHub API whose attempt times out after 15 seconds,
// and that budget is shared with the request for an installation access token and the actual request.
// So this must be well under that timeout.
const kmsSignTimeout = 5 * time.Second

// KMSAPI is the subset of the AWS KMS API client used to sign JWTs.
type KMSAPI interface {
	Sign(ctx context.Context, params *kms.SignInput, optFns ...func(*kms.Options)) (*kms.SignOutput, error)
}

// KMSSigner signs JWTs for a GitHub App by delegating the signing to AWS KMS.
// The GitHub App's private key is imported into AWS KMS, so the app never has the private key.
// It implements ghinstallation.Signer.
type KMSSigner struct {
	method *kmsSigningMethod
}

// NewKMSSigner creates a KMSSigner with the default AWS config.
// keyID is a key ID, key ARN, alias name, or alias ARN.
// If keyID is an ARN, the region of the ARN is used.
// Otherwise, the region is resolved from the default AWS config, such as the environment variable AWS_REGION.
func NewKMSSigner(ctx context.Context, keyID string) (*KMSSigner, error) {
	cfg, err := NewConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	if a, err := arn.Parse(keyID); err == nil && a.Region != "" {
		cfg.Region = a.Region
	}
	return NewKMSSignerWithClient(kms.NewFromConfig(cfg), keyID), nil
}

// NewKMSSignerWithClient creates a KMSSigner with the given AWS KMS API client.
func NewKMSSignerWithClient(client KMSAPI, keyID string) *KMSSigner {
	return &KMSSigner{
		method: &kmsSigningMethod{
			client:  client,
			keyID:   keyID,
			timeout: kmsSignTimeout,
		},
	}
}

// Sign signs the given claims with RS256 using AWS KMS and returns a JWT.
func (s *KMSSigner) Sign(claims jwt.Claims) (string, error) {
	// The key is passed to kmsSigningMethod.Sign, but it's unused because the private key is in AWS KMS.
	token, err := jwt.NewWithClaims(s.method, claims).SignedString(nil)
	if err != nil {
		return "", fmt.Errorf("sign a JWT with AWS KMS: %w", err)
	}
	return token, nil
}

// kmsSigningMethod is a jwt.SigningMethod for RS256 that signs with AWS KMS.
// The key must be an RSA_2048 key whose key usage is SIGN_VERIFY.
type kmsSigningMethod struct {
	client  KMSAPI
	keyID   string
	timeout time.Duration
}

func (m *kmsSigningMethod) Alg() string {
	return jwt.SigningMethodRS256.Alg()
}

func (m *kmsSigningMethod) Verify(_, _ string, _ any) error {
	return errors.New("verifying a JWT isn't supported")
}

// Sign signs the signing string and returns the base64url encoded signature.
// The interface of ghinstallation.Signer doesn't accept a context, so the timeout is set here.
func (m *kmsSigningMethod) Sign(signingString string, _ any) (string, error) {
	digest := sha256.Sum256([]byte(signingString))
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	out, err := m.client.Sign(ctx, &kms.SignInput{
		KeyId:            aws.String(m.keyID),
		Message:          digest[:],
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecRsassaPkcs1V15Sha256,
	})
	if err != nil {
		return "", fmt.Errorf("call kms:Sign: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(out.Signature), nil
}
