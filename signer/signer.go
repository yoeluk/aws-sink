package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/yoeluk/aws-sink/aws"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type CanonRequest struct {
	// aws
	Creds          *aws.Credentials
	Region         string
	Service        string
	VersionRequest string

	// V4 data
	httpVerb    string
	date        string
	queryParams map[string]string
	amzHeaders  map[string]string
	resource    string
}

func (c *CanonRequest) RequestString() string {
	qstring := canonString(c.queryParams, "=", "&", true)
	headers := canonString(c.amzHeaders, ":", "\n", false)
	keys := strings.Join(sortedKeys(c.amzHeaders), ";")
	contentSha256 := c.amzHeaders["x-amz-content-sha256"]
	return fmt.Sprintf("%s\n%s\n%s\n%s\n\n%s\n%s", c.httpVerb, c.resource, qstring, headers, keys, contentSha256)
}

func (c *CanonRequest) StringToSignV4() string {
	sha := sha256.New()
	sha.Write([]byte(c.RequestString()))
	canRequest := sha.Sum(nil)
	date := c.date
	if awsDate, ok := c.amzHeaders["x-amz-date"]; ok {
		date = awsDate
	}
	return "AWS4-HMAC-SHA256" + "\n" +
		date + "\n" +
		date[:8] + "/" + c.Region + "/" + c.Service + "/" + c.VersionRequest + "\n" +
		hex.EncodeToString(canRequest)
}

func (c *CanonRequest) SignatureV4() string {
	date := c.date
	if awsDate, ok := c.amzHeaders["x-amz-date"]; ok {
		date = awsDate
	}
	dateKey := hmac.New(sha256.New, []byte("AWS4"+c.Creds.AccessSecretKey))
	dateKey.Write([]byte(date[:8]))
	dateRegionKey := hmac.New(sha256.New, dateKey.Sum(nil))
	dateRegionKey.Write([]byte(c.Region))
	dateRegionServiceKey := hmac.New(sha256.New, dateRegionKey.Sum(nil))
	dateRegionServiceKey.Write([]byte(c.Service))
	signingKey := hmac.New(sha256.New, dateRegionServiceKey.Sum(nil))
	signingKey.Write([]byte("aws4_request"))
	signatureV4 := hmac.New(sha256.New, signingKey.Sum(nil))
	signatureV4.Write([]byte(c.StringToSignV4()))
	return hex.EncodeToString(signatureV4.Sum(nil))
}

func (c *CanonRequest) AuthHeader() string {
	date := c.date
	if awsDate, ok := c.amzHeaders["x-amz-date"]; ok {
		date = awsDate
	}
	return "AWS4-HMAC-SHA256 " +
		"Credential=" + c.Creds.AccessKeyId + "/" + date[:8] + "/" + c.Region + "/" + c.Service + "/" + c.VersionRequest +
		",SignedHeaders=" + strings.Join(sortedKeys(c.amzHeaders), ";") +
		",Signature=" + c.SignatureV4()
}

func Signer(request *http.Request, payload []byte, crTemplate CanonRequest) *CanonRequest {
	now := time.Now()
	formatted := strings.ReplaceAll(
		strings.ReplaceAll(now.UTC().Format(time.RFC3339), "-", ""),
		":", "")
	awsDate := formatted[:len(formatted)-3] + "00Z"
	if date := request.Header.Get("date"); date == "" {
		request.Header.Set("date", now.Local().Format(time.RFC1123))
	}
	sha := sha256.New()
	sha.Write(payload)
	request.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha.Sum(nil)))
	request.Header.Set("X-Amz-Date", awsDate)
	if crTemplate.Creds.SecurityToken != "" {
		request.Header.Set("x-amz-security-token", crTemplate.Creds.SecurityToken)
	}
	return updateCanonRequest(request, &crTemplate)
}

func updateCanonRequest(req *http.Request, cr *CanonRequest) *CanonRequest {
	m := map[string][]string(req.Header)
	headers := make(map[string]string, len(m))
	for k, vs := range m {
		headers[strings.ToLower(k)] = strings.TrimSpace(strings.Join(vs, ","))
	}
	cr.httpVerb = req.Method
	cr.date = headers["date"]
	cr.amzHeaders = headers
	if req.URL.Path != "" {
		cr.resource = strings.TrimSpace(req.URL.Path)
	}
	return cr
}

func canonString(in map[string]string, sep string, inter string, encoding bool) string {
	var c string
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if c != "" {
			c = c + inter
		}
		if encoding {
			c = c + fmt.Sprintf("%s%s%s", url.QueryEscape(k), sep, url.QueryEscape(in[k]))
		} else {
			c = c + fmt.Sprintf("%s%s%s", k, sep, in[k])
		}
	}
	return c
}

func sortedKeys(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)
	return keys
}
