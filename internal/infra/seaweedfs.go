package infra

import (
	"bytes"
	c "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SeaWeedFS struct {
	hostPort string
	client   *http.Client
}

func NewSeaWeedFS(hostPort string) *SeaWeedFS {
	return &SeaWeedFS{hostPort: hostPort, client: &http.Client{}}
}

type restHelper struct {
	url     url.URL
	method  string
	payload []byte
	status  int
}

func (nh restHelper) run(ctx c.Context, cl *http.Client) ([]byte, error) {
	if nh.url.Scheme == "" {
		nh.url.Scheme = "http"
	}

	if nh.method == "" {
		nh.method = "GET"
	}

	var r io.Reader

	if nh.payload != nil {
		r = bytes.NewReader(nh.payload)
	}

	rqst, err := http.NewRequestWithContext(ctx, nh.method, nh.url.String(), r)

	if err != nil {
		return nil, err
	}

	if nh.payload != nil {
		rqst.Header.Set("Content-Type", "application/octet-stream")
	}

	if nh.status == 0 {
		nh.status = http.StatusOK
	}

	rsps, err := cl.Do(rqst)

	if err != nil {
		return nil, err
	}

	defer rsps.Body.Close()

	if rsps.StatusCode != nh.status {
		return nil, fmt.Errorf(
			"status code %d (%s)",
			rsps.StatusCode, http.StatusText(rsps.StatusCode),
		)
	}

	body, err := io.ReadAll(rsps.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *SeaWeedFS) Create(ctx c.Context, blob []byte) (string, error) {
	body, err := restHelper{
		url: url.URL{
			Host: s.hostPort,
			Path: "/dir/assign",
		},
		method:  "POST",
		payload: nil,
		status:  http.StatusOK,
	}.run(ctx, s.client)

	if err != nil {
		return "", err
	}

	var unpacked struct {
		Fid       string `json:"fid"`
		Url       string `json:"url"`
		PublicUrl string `json:"publicUrl"`
		Count     int    `json:"count"`
	}

	err = json.Unmarshal(body, &unpacked)

	if err != nil {
		return "", err
	}

	_, err = restHelper{
		url: url.URL{
			Host: unpacked.Url,
			Path: unpacked.Fid,
		},
		method:  "POST",
		payload: blob,
		status:  http.StatusCreated,
	}.run(ctx, s.client)

	if err != nil {
		return "", err
	}

	return unpacked.Fid, nil
}

func (s *SeaWeedFS) Read(ctx c.Context, id string) ([]byte, error) {
	return restHelper{
		url: url.URL{
			Scheme: "http",
			Host:   s.hostPort,
			Path:   id,
		},
		method: "GET",
	}.run(ctx, s.client)
}

func (s *SeaWeedFS) Delete(ctx c.Context, id string) error {
	_, err := restHelper{
		url: url.URL{
			Scheme: "http",
			Host:   s.hostPort,
			Path:   id,
		},
		method:  "DELETE",
		payload: nil,
		status:  http.StatusAccepted,
	}.run(ctx, s.client)

	return err
}
