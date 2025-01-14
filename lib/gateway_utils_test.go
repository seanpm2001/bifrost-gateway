package lib

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	ifacepath "github.com/ipfs/boxo/coreiface/path"
	"github.com/ipfs/boxo/gateway"
)

func TestContentPathToCarUrl(t *testing.T) {
	negativeOffset := int64(-42)
	testCases := []struct {
		contentPath string // to be turned into gateway.ImmutablePath
		carParams   gateway.CarParams
		expectedUrl string // url.URL.String()
	}{
		{
			contentPath: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			carParams:   gateway.CarParams{},
			expectedUrl: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi?format=car",
		},
		{
			contentPath: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			carParams:   gateway.CarParams{Scope: "entity", Range: &gateway.DagByteRange{From: 0, To: nil}},
			expectedUrl: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi?format=car&dag-scope=entity&entity-bytes=0:*",
		},
		{
			contentPath: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			carParams:   gateway.CarParams{Scope: "block"},
			expectedUrl: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi?format=car&dag-scope=block",
		},
		{
			contentPath: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			carParams:   gateway.CarParams{Scope: "entity", Range: &gateway.DagByteRange{From: 4, To: &negativeOffset}},
			expectedUrl: "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi?format=car&dag-scope=entity&entity-bytes=4:-42",
		},
		{
			// a regression test for case described in https://github.com/ipfs/gateway-conformance/issues/115
			contentPath: "/ipfs/bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjze/I/Auditorio_de_Tenerife%2C_Santa_Cruz_de_Tenerife%2C_España%2C_2012-12-15%2C_DD_02.jpg.webp",
			carParams:   gateway.CarParams{Scope: "entity", Range: &gateway.DagByteRange{From: 0, To: nil}},
			expectedUrl: "/ipfs/bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjze/I/Auditorio_de_Tenerife%252C_Santa_Cruz_de_Tenerife%252C_Espa%C3%B1a%252C_2012-12-15%252C_DD_02.jpg.webp?format=car&dag-scope=entity&entity-bytes=0:*",
		},
	}

	for _, tc := range testCases {
		t.Run("TestContentPathToCarUrl", func(t *testing.T) {
			contentPath, err := gateway.NewImmutablePath(ifacepath.New(tc.contentPath))
			if err != nil {
				t.Fatal(err)
			}
			result := contentPathToCarUrl(contentPath, tc.carParams).String()
			if result != tc.expectedUrl {
				t.Errorf("Expected %q, but got %q", tc.expectedUrl, result)
			}
		})
	}
}

type testErr struct {
	message    string
	retryAfter time.Duration
}

func (e *testErr) Error() string {
	return e.message
}

func (e *testErr) RetryAfter() time.Duration {
	return e.retryAfter
}

func TestGatewayErrorRetryAfter(t *testing.T) {
	originalErr := &testErr{message: "test", retryAfter: time.Minute}
	var (
		convertedErr error
		gatewayErr   *gateway.ErrorRetryAfter
	)

	// Test unwrapped
	convertedErr = GatewayError(originalErr)
	ok := errors.As(convertedErr, &gatewayErr)
	assert.True(t, ok)
	assert.EqualValues(t, originalErr.retryAfter, gatewayErr.RetryAfter)

	// Test wrapped.
	convertedErr = GatewayError(fmt.Errorf("wrapped error: %w", originalErr))
	ok = errors.As(convertedErr, &gatewayErr)
	assert.True(t, ok)
	assert.EqualValues(t, originalErr.retryAfter, gatewayErr.RetryAfter)
}
