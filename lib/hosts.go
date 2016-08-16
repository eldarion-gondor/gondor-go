package gondor

import (
	"fmt"
	"net/url"
)

type HostNameResource struct {
	client *Client
}

type HostName struct {
	Instance *string `json:"instance,omitempty"`
	Host     *string `json:"host,omitempty"`

	URL *string `json:"url,omitempty"`

	r *HostNameResource
}

func (r *HostNameResource) findOne(url *url.URL) (*HostName, error) {
	var res *HostName
	resp, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("host not found")
	}
	res.r = r
	return res, nil
}

func (r *HostNameResource) Create(hostName *HostName) error {
	url := r.client.buildBaseURL("hosts/")
	_, err := r.client.Post(url, hostName, hostName)
	if err != nil {
		return err
	}
	return nil
}

func (r *HostNameResource) GetFromURL(value string) (*HostName, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return r.findOne(u)
}

func (r *HostNameResource) Get(instanceURL string, host string) (*HostName, error) {
	url := r.client.buildBaseURL("hosts/find/")
	q := url.Query()
	q.Set("instance", instanceURL)
	q.Set("host", host)
	url.RawQuery = q.Encode()
	return r.findOne(url)
}

func (r *HostNameResource) List(instanceURL *string) ([]*HostName, error) {
	url := r.client.buildBaseURL("hosts/")
	q := url.Query()
	if instanceURL != nil {
		q.Set("instance", *instanceURL)
	}
	url.RawQuery = q.Encode()
	var res []*HostName
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *HostNameResource) Update(hostName HostName) error {
	u, _ := url.Parse(*hostName.URL)
	hostName.URL = nil
	_, err := r.client.Patch(u, &hostName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *HostNameResource) Delete(hostName *HostName) error {
	hostNames, err := r.List(hostName.Instance)
	if err != nil {
		return err
	}
	var foundHostName *HostName
	for i := range hostNames {
		foundHostName = hostNames[i]
		if hostName.Host == foundHostName.Host {
			break
		}
	}
	u, _ := url.Parse(*foundHostName.URL)
	_, err = r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}
