package aws

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
)

// fakeKMS signs a digest with a local RSA key in the same way as AWS KMS.
type fakeKMS struct {
	key   *rsa.PrivateKey
	err   error
	input *kms.SignInput
}

func (f *fakeKMS) Sign(_ context.Context, params *kms.SignInput, _ ...func(*kms.Options)) (*kms.SignOutput, error) {
	f.input = params
	if f.err != nil {
		return nil, f.err
	}
	sig, err := rsa.SignPKCS1v15(nil, f.key, crypto.SHA256, params.Message)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return &kms.SignOutput{Signature: sig}, nil
}

func TestKMSSigner_Sign(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeKMS{key: key}
	keyID := "arn:aws:kms:ap-northeast-1:123456789012:key/xxx"
	signer := NewKMSSignerWithClient(client, keyID)

	now := time.Now().Truncate(time.Second)
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		Issuer:    strconv.Itoa(123),
	}
	token, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(keyID, *client.input.KeyId); diff != "" {
		t.Errorf("KeyId (-want +got):\n%s", diff)
	}
	if client.input.MessageType != types.MessageTypeDigest {
		t.Errorf("MessageType = %s, want DIGEST", client.input.MessageType)
	}
	if client.input.SigningAlgorithm != types.SigningAlgorithmSpecRsassaPkcs1V15Sha256 {
		t.Errorf("SigningAlgorithm = %s, want RSASSA_PKCS1_V1_5_SHA_256", client.input.SigningAlgorithm)
	}

	// The JWT must be verifiable with the public key as a standard RS256 JWT.
	parsed := &jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(token, parsed, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != "RS256" {
			t.Errorf("alg = %s, want RS256", token.Method.Alg())
		}
		return &key.PublicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"})); err != nil {
		t.Fatalf("verify the JWT: %v", err)
	}
	if diff := cmp.Diff(claims, parsed); diff != "" {
		t.Errorf("claims (-want +got):\n%s", diff)
	}
}

func TestKMSSigner_Sign_error(t *testing.T) {
	t.Parallel()
	kmsErr := errors.New("access denied")
	signer := NewKMSSignerWithClient(&fakeKMS{err: kmsErr}, "alias/github-app")
	if _, err := signer.Sign(&jwt.RegisteredClaims{}); !errors.Is(err, kmsErr) {
		t.Errorf("error = %v, want %v", err, kmsErr)
	}
}
