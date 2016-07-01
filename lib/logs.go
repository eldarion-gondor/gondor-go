package gondor

import "strconv"

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

func (r *LogResource) ListByInstance(instanceURL string, lines int, since string) ([]*LogRecord, error) {
	url := r.client.buildBaseURL("logs/")
	q := url.Query()
	q.Add("instance", instanceURL)
	if lines > 0 {
		q.Add("size", strconv.Itoa(lines))
	}
	if since != "" {
		q.Add("since", since)
	}
	url.RawQuery = q.Encode()
	var res []*LogRecord
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *LogResource) ListByService(serviceURL string, lines int, since string) ([]*LogRecord, error) {
	url := r.client.buildBaseURL("logs/")
	q := url.Query()
	q.Add("service", serviceURL)
	if lines > 0 {
		q.Add("size", strconv.Itoa(lines))
	}
	if since != "" {
		q.Add("since", since)
	}
	url.RawQuery = q.Encode()
	var res []*LogRecord
	_, err := r.client.Get(url, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
