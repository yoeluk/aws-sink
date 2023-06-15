package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/yoeluk/aws-sink/aws"
	"github.com/yoeluk/aws-sink/log"
	"github.com/yoeluk/aws-sink/signer"
	"io"
	"net/http"
	"time"
)

type Sink struct {
	client     *http.Client
	template   *signer.CanonRequest
	bucketHost string
	prefix     string
	timeout    int
}

func New(bucket, prefix, region string, timeout int, creds *aws.Credentials) *Sink {
	cr := &signer.CanonRequest{
		Creds:          creds,
		Region:         region,
		Service:        "s3",
		VersionRequest: "aws4_request",
	}
	return &Sink{
		client:     &http.Client{},
		template:   cr,
		bucketHost: fmt.Sprintf("https://%s.s3.amazonaws.com", bucket),
		prefix:     prefix,
		timeout:    timeout,
	}
}

func (s *Sink) Put(name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	request, err := http.NewRequest("PUT", s.bucketHost+s.prefix+"/"+name, bytes.NewReader(payload))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
	if cancel != nil {
		defer cancel()
	}
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Host", request.URL.Host)
	cr := signer.Signer(request, payload, *s.template)
	request.Header.Set("Authorization", cr.AuthHeader())
	resp, err := s.client.Do(request.WithContext(ctx))
	if err != nil {
		log.Error(fmt.Sprintf("found an error putting object %q, status: %q, error: %s", name, resp.Status, err.Error()))
		return nil, err
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf(cr.RequestString())
	}
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("there was an error reading the S3's response: %q", err.Error()))
	}
	copyHeader(rw.Header(), resp.Header)
	return response, nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
