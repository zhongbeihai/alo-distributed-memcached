package pkg

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/alo-distributed-memcached/pb"
	consistenthash "github.com/alo-distributed-memcached/pkg/consistent_hash"
	"google.golang.org/protobuf/proto"
)

const (
	defaultBasePath = "/alo-cache/"
	defaultReplicas = 50
)

/*
1. HTTPPool implements PeerPicker interface.
When the data is not in current node, current node will use HTTPPool.PickPeer()
to get the **HTTPGetter** of other node (not the other node) that has the data.
2. HTTPPool implements ServerHTTP interface.
It will handle the request from other node and return the data to the other node.
*/
type HTTPPool struct {
	self       string // store it owns hostname and port
	basePath   string
	mu         sync.Mutex
	peers      *consistenthash.ConsistentHashMap
	httpGetter map[string]*HTTPGetter // keyed by e.g. "http://10.0.0.2:8008"
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (h *HTTPPool) SetPeers(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.peers = consistenthash.NewConsistentHashMap(defaultReplicas, nil)
	h.peers.AddNode(peers...)
	h.httpGetter = make(map[string]*HTTPGetter)

	for _, peer := range peers {
		h.httpGetter[peer] = &HTTPGetter{baseURL: peer + h.basePath}
	}
}

// ------ PeerPicker interface ------
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

var _ PeerPicker = (*HTTPPool)(nil)

// implement PeerPicker interface
// PickPeer() is used to pick a peer from the consistent hash ring to get the data from other node
// return PeerGetter, then should call PeerGetter.GetDataFromPeer() to get the data
func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if peer := h.peers.GetNode(key); peer != "" && peer != h.self {
		h.log("pick peer %s", peer)
		return h.httpGetter[peer], true
	}
	return nil, false
}

func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("unexpected path:" + r.URL.Path)
	}
	h.log("%s, %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "error in request url", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group", http.StatusBadRequest)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := proto.Marshal(&pb.Response{
		Value: view.ByteSlice(),
	})
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

func (h *HTTPPool) log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}
