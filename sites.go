package gondor

import (
	"errors"
	"fmt"
	"net/url"
)

type SiteResource struct {
	client *Client
}

type Site struct {
	Name          string         `json:"name,omitempty"`
	Key           string         `json:"key,omitempty"`
	ResourceGroup *ResourceGroup `json:"resource_group,omitempty"`
	Instances     []Instance     `json:"instances,omitempty"`
	Users         []struct {
		User struct {
			Username string `json:"username,omitempty"`
		} `json:"user,omitempty"`
		Role string `json:"role,omitempty"`
	} `json:"users,omitempty"`

	URL string `json:"url,omitempty"`

	r *SiteResource
}

func (r *SiteResource) Create(site *Site) error {
	url := r.client.buildBaseURL("sites/")
	_, err := r.client.Post(url, site, site)
	if err != nil {
		return err
	}
	return nil
}

func (r *SiteResource) List(resourceGroup *ResourceGroup) ([]*Site, error) {
	url := r.client.buildBaseURL("sites/")
	q := url.Query()
	if resourceGroup != nil {
		q.Set("resource_group", resourceGroup.URL)
	}
	url.RawQuery = q.Encode()
	var res []*Site
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].r = r
	}
	return res, nil
}

func (r *SiteResource) findOne(url *url.URL) (*Site, error) {
	var res *Site
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	res.r = r
	return res, nil
}

func (r *SiteResource) Get(name string, resourceGroup *ResourceGroup) (*Site, error) {
	url := r.client.buildBaseURL("sites/find/")
	q := url.Query()
	q.Set("name", name)
	if resourceGroup != nil {
		q.Set("resource_group", resourceGroup.URL)
	}
	url.RawQuery = q.Encode()
	site, err := r.findOne(url)
	if _, ok := err.(ErrNotFound); ok {
		msg := fmt.Sprintf("site %q was not found", name)
		if resourceGroup != nil {
			msg += fmt.Sprintf(" in resource group %q", resourceGroup.Name)
		}
		return site, fmt.Errorf(msg)
	}
	return site, err
}

func (r *SiteResource) Delete(site *Site) error {
	if site.URL == "" {
		return errors.New("missing site URL")
	}
	u, _ := url.Parse(site.URL)
	_, err := r.client.Delete(u, nil)
	if err != nil {
		return err
	}
	return nil
}

func (site *Site) AddUser(email string, role string) error {
	url := site.r.client.buildBaseURL("site_users/")
	req := struct {
		Site  *Site  `json:"site,omitempty"`
		Email string `json:"email,omitempty"`
		Role  string `json:"role,omitempty"`
	}{
		Site:  site,
		Email: email,
		Role:  role,
	}
	_, err := site.r.client.Post(url, &req, nil)
	if err != nil {
		return err
	}
	return nil
}
