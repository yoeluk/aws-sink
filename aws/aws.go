package aws

import (
	"encoding/json"
	"github.com/yoeluk/aws-sink/log"
	"io"
	"net/http"
	"os"
	"time"
)

const onErrorRotationInterval = 10 * time.Second

type Credentials struct {
	AccessKeyId     string    `json:"AccessKeyId"`
	AccessSecretKey string    `json:"SecretAccessKey"`
	SecurityToken   string    `json:"Token"`
	RoleArn         string    `json:"RoleArn"`
	Expiration      time.Time `json:"Expiration"`
}

func EcsCredentials() *Credentials {
	creds := &Credentials{}
	go ecsCredentials(creds)
	return creds
}

func ecsCredentials(creds *Credentials) {
	client := &http.Client{}
	credsURL := "http://169.254.170.2" + os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	for {
		err := refreshEcsCredentials(creds, client, credsURL)
		if err != nil {
			log.Error(err.Error())
			onErrorTimer := time.NewTimer(onErrorRotationInterval)
			<-onErrorTimer.C
		} else {
			duration := time.Until(creds.Expiration)
			d := duration / 2
			renewalTimer := time.NewTimer(d)
			<-renewalTimer.C
		}
	}
}

func refreshEcsCredentials(creds *Credentials, client *http.Client, credsURL string) error {
	credReq, err := http.NewRequest("GET", credsURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(credReq)
	if err != nil {
		return err
	}
	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rb, creds)
	if err != nil {
		return err
	}
	return nil
}
