package gondor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type BuildResource struct {
	client *Client
}

type Build struct {
	Site         *string `json:"site,omitempty"`
	Instance     *string `json:"instance,omitempty"`
	Label        *string `json:"label,omitempty"`
	Ref          *string `json:"ref,omitempty"`
	Sha          *string `json:"sha,omitempty"`
	BuildpackURL *string `json:"buildpack_url,omitempty"`
	Creator      *string `json:"creator,omitempty"`
	Created      *string `json:"created,omitempty"`

	URL *string `json:"url,omitempty"`

	r *BuildResource
}

func (r *BuildResource) List(siteURL *string, instanceURL *string, limit int) ([]*Build, error) {
	url := r.client.buildBaseURL("builds/")
	q := url.Query()
	if siteURL != nil {
		q.Set("site", *siteURL)
	}
	if instanceURL != nil {
		q.Set("instance", *instanceURL)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	url.RawQuery = q.Encode()
	var res []*Build
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *BuildResource) Create(build *Build) error {
	url := r.client.buildBaseURL("builds/")
	_, err := r.client.Post(url, build, build)
	if err != nil {
		return err
	}
	build.r = r
	return nil
}

func (build *Build) Perform(blob io.Reader) (string, error) {
	// buffer blob to disk
	file, err := ioutil.TempFile("", "blob-")
	if err != nil {
		return "", err
	}
	defer file.Close()
	defer os.Remove(file.Name())
	if _, err := io.Copy(file, blob); err != nil {
		return "", err
	}

	// make request to build to perform it
	blobFile, err := os.Open(file.Name())
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("PUT", *build.URL, blobFile)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", build.r.client.cfg.Auth.AccessToken))
	req.Header.Add("Content-Type", "application/x-tar")
	req.Header.Add("Content-Disposition", "attachment; filename=blob.tar")
	fi, err := blobFile.Stat()
	if err != nil {
		return "", err
	}
	req.ContentLength = fi.Size()
	resp, err := build.r.client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload struct {
		Endpoint string `json:"endpoint,omitempty"`
	}
	if resp.StatusCode < 300 {
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("build: non-200 response; got %s", resp.Status)
	}
	return payload.Endpoint, nil
}
