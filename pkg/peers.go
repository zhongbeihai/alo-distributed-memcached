package pkg

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	GetDataFromPeer(group string, key string) ([]byte, error)
}

type HTTPGetter struct {
	baseURL string
}

func (h *HTTPGetter) GetDataFromPeer(group string, key string) ([]byte, error) {
	// /<basepath>/<groupname>/<key>
	url := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetDataFromPeer():server returner %v", response.Status)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("GetDataFromPeer():reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*HTTPGetter)(nil)
