package gondor

import (
	"net/url"
	"strconv"
	"time"
)

type LogResource struct {
	client *Client
}

type LogRecord struct {
	ID        *string `json:"id"`
	Timestamp *string `json:"@timestamp"`
	Message   *string `json:"log"`
	Stream    *string `json:"stream"`
	Tag       *string `json:"tag"`
}

// LogRequestOpts ...
type LogRequestOpts struct {
	PageSize  int
	After     *time.Time
	Before    *time.Time
	PageToken string
	Order     string
}

// LogRecordPage represents a single log record page
type LogRecordPage struct {
	Records       []*LogRecord
	NextPageToken string
}

func (r *LogResource) query(u *url.URL, q url.Values, opts LogRequestOpts) (*LogRecordPage, error) {
	if opts.PageSize > 0 {
		q.Add("size", strconv.Itoa(opts.PageSize))
	}
	if opts.After != nil {
		q.Add("after", opts.After.Format("2006-01-02T15:04:05-0700"))
	}
	if opts.Before != nil {
		q.Add("before", opts.Before.Format("2006-01-02T15:04:05-0700"))
	}
	if opts.PageToken != "" {
		q.Add("page_token", opts.PageToken)
	}
	u.RawQuery = q.Encode()
	var res []*LogRecord
	resp, err := r.client.Get(u, &res)
	if err != nil {
		return nil, err
	}
	page := &LogRecordPage{
		Records:       res,
		NextPageToken: resp.Header.Get("X-Log-Page-Token"),
	}
	return page, nil
}

// ListByInstance ...
func (r *LogResource) ListByInstance(instanceURL string, opts LogRequestOpts) (*LogRecordPage, error) {
	url := r.client.buildBaseURL("logs/")
	q := url.Query()
	q.Add("instance", instanceURL)
	return r.query(url, q, opts)
}

// ListByService ...
func (r *LogResource) ListByService(serviceURL string, opts LogRequestOpts) (*LogRecordPage, error) {
	url := r.client.buildBaseURL("logs/")
	q := url.Query()
	q.Add("service", serviceURL)
	return r.query(url, q, opts)
}
