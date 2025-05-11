package pkg

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/alo-distributed-memcached/pb"
	"google.golang.org/protobuf/proto"
)


type PeerGetter interface {
	GetDataFromPeer(in *pb.Request, out *pb.Response) error
}

/*
HTTPGetter implements PeerGetter interface and uses HTTP to get data from other nodes.
Implement HTTP Client function for Node
*/
type HTTPGetter struct {
	baseURL string
}

func (h *HTTPGetter) GetDataFromPeer(in *pb.Request, out *pb.Response) error {
	// /<basepath>/<groupname>/<key>
	url := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GetDataFromPeer():server returner %v", response.Status)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("GetDataFromPeer():reading response body: %v", err)
	}

	if err := proto.Unmarshal(bytes, out); err != nil{
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ PeerGetter = (*HTTPGetter)(nil)
