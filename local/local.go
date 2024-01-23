package local

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/yoeluk/aws-sink/aws"
	"github.com/yoeluk/aws-sink/log"
	"github.com/yoeluk/aws-sink/signer"
)

type Sink struct {
	client         *http.Client
	template       *signer.CanonRequest
	localDirectory string
}

func New(region, localDirectory string, creds *aws.Credentials) *Sink {
	cr := &signer.CanonRequest{
		Creds:          creds,
		Region:         region,
		Service:        "local",
		VersionRequest: "aws4_request",
	}
	return &Sink{
		client:         &http.Client{},
		template:       cr,
		localDirectory: localDirectory,
	}
}

func (s *Sink) Put(name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", s.localDirectory, name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	_, err = f.Write(payload)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debug(fmt.Sprintf("put %q object in %q local directory.", name, s.localDirectory))
	return []byte(fmt.Sprintf("object %q was put in the local sink", name)), nil
}

func (s *Sink) Post(path string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	return s.Put(path+"/"+uuid.NewString(), payload, contentType, rw)
}
