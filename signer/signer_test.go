package signer

import (
	"bytes"
	"fmt"
	"github.com/yoeluk/aws-sink/aws"
	"net/http"
	"testing"
)

func TestCanonRequestV4(t *testing.T) {
	testCases := []struct {
		sig           *CanonRequest
		expectedSig   string
		expectedAuthZ string
		name          string
	}{
		{
			name:          "get object request",
			expectedSig:   "f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41",
			expectedAuthZ: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41",
			sig: &CanonRequest{
				Creds: &aws.Credentials{
					AccessSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
				},
				httpVerb:       "GET",
				date:           "20130524T000000Z",
				Region:         "us-east-1",
				Service:        "s3",
				resource:       "/test.txt",
				VersionRequest: "aws4_request",
				amzHeaders: map[string]string{
					"host":                 "examplebucket.s3.amazonaws.com",
					"range":                "bytes=0-9",
					"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					"x-amz-date":           "20130524T000000Z",
				},
			},
		},
		{
			name:          "put object request",
			expectedSig:   "98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd",
			expectedAuthZ: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class,Signature=98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd",
			sig: &CanonRequest{
				Creds: &aws.Credentials{
					AccessSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
				},
				httpVerb:       "PUT",
				date:           "Fri, 24 May 2013 00:00:00 GMT",
				Region:         "us-east-1",
				Service:        "s3",
				resource:       "/test%24file.text",
				VersionRequest: "aws4_request",
				amzHeaders: map[string]string{
					"host":                 "examplebucket.s3.amazonaws.com",
					"date":                 "Fri, 24 May 2013 00:00:00 GMT",
					"x-amz-content-sha256": "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072",
					"x-amz-date":           "20130524T000000Z",
					"x-amz-storage-class":  "REDUCED_REDUNDANCY",
				},
			},
		},
	}

	for _, tt := range testCases {
		signature := tt.sig.SignatureV4()
		authz := tt.sig.AuthHeader()
		if signature != tt.expectedSig {
			t.Errorf("found an error while testing %s: the signature didn't match the expected string\nexpected: %s\nfound: %s\nStringToSign:\n%s\nRequestString: \n%s\n", tt.name, tt.expectedSig, signature, tt.sig.StringToSignV4(), tt.sig.RequestString())
		}
		if authz != tt.expectedAuthZ {
			t.Errorf("found unexpected authorization header; expected:\n%s,\nfound:\n%s", tt.expectedAuthZ, authz)
		}
	}
}

func TestSignerV4(t *testing.T) {
	crTemplate := &CanonRequest{
		Creds: &aws.Credentials{
			AccessSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		},
		Region:         "us-east-1",
		Service:        "s3",
		VersionRequest: "aws4_request",
	}

	r1, _ := http.NewRequest("PUT", "https://examplebucket.s3.amazonaws.com/text/test/file.text", bytes.NewReader([]byte("")))

	testCases := []struct {
		name     string
		expected string
		request  *http.Request
	}{
		{
			name:     "basic",
			expected: "hello",
			request:  r1,
		},
	}

	for _, tt := range testCases {
		r1.Header.Set("Content-Type", "application/json")
		r1.Header.Set("Host", r1.URL.Host)
		cr := Signer(tt.request, make([]byte, 0), *crTemplate)
		fmt.Printf("%v\n", cr.AuthHeader())
		fmt.Printf("%v\n", cr.StringToSignV4())
		fmt.Printf("%v\n", cr.RequestString())
	}
}
