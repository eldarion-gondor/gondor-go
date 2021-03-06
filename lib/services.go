package gondor

import (
	"fmt"
	"net/url"
	"strings"
)

type ServiceResource struct {
	client *Client
}

type Service struct {
	Instance *string           `json:"instance,omitempty"`
	Name     *string           `json:"name,omitempty"`
	Kind     *string           `json:"kind,omitempty"`
	Image    *string           `json:"image,omitempty"`
	Size     *string           `json:"size,omitempty"`
	Replicas *int              `json:"replicas,omitempty"`
	Network  *string           `json:"network,omitempty"`
	Volume   *string           `json:"volume,omitempty"`
	State    *string           `json:"state,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	WebURL   *string           `json:"web_url,omitempty"`

	// create only
	Version   *string `json:"version,omitempty"`
	OpenPorts string  `json:"open_ports,omitempty"`

	// update only
	DesiredState    *string `json:"desired_state,omitempty"`
	DesiredReplicas *int    `json:"desired_replicas,omitempty"`

	URL *string `json:"url,omitempty"`

	r *ServiceResource
}

func (r *ServiceResource) findOne(url *url.URL) (*Service, error) {
	var res *Service
	resp, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("service not found")
	}
	res.r = r
	return res, nil
}

func (r *ServiceResource) Create(service *Service) error {
	url := r.client.buildBaseURL("services/")
	_, err := r.client.Post(url, service, service)
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceResource) GetFromURL(value string) (*Service, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return r.findOne(u)
}

func (r *ServiceResource) Get(instanceURL string, name string) (*Service, error) {
	url := r.client.buildBaseURL("services/find/")
	q := url.Query()
	q.Set("instance", instanceURL)
	q.Set("name", name)
	url.RawQuery = q.Encode()
	return r.findOne(url)
}

func (r *ServiceResource) List(instanceURL *string) ([]*Service, error) {
	url := r.client.buildBaseURL("services/")
	q := url.Query()
	if instanceURL != nil {
		q.Set("instance", *instanceURL)
	}
	url.RawQuery = q.Encode()
	var res []*Service
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *ServiceResource) Update(service Service) error {
	u, _ := url.Parse(*service.URL)
	service.URL = nil
	_, err := r.client.Patch(u, &service, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *ServiceResource) Delete(serviceURL string) error {
	u, _ := url.Parse(serviceURL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Restart() error {
	return s.SetState("restarted")
}

func (s *Service) SetState(state string) error {
	desiredService := Service{
		DesiredState: &state,
	}
	u, _ := url.Parse(*s.URL)
	_, err := s.r.client.Patch(u, &desiredService, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SetReplicas(n int) error {
	desiredService := Service{
		DesiredReplicas: &n,
	}
	u, _ := url.Parse(*s.URL)
	_, err := s.r.client.Patch(u, &desiredService, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Run(cmd []string, size string) (string, error) {
	u, _ := url.Parse(*s.URL + "run/")
	up := struct {
		Command string `json:"command,omitempty"`
		Size *string `json:"size,omitempty"`
	}{
		Command: strings.Join(cmd, " "),
	}
	if size != "" {
		up.Size = &size
	}
	down := struct {
		Endpoint string `json:"endpoint"`
	}{}
	_, err := s.r.client.Post(u, &up, &down)
	if err != nil {
		return "", err
	}
	return down.Endpoint, nil
}
