package uiattr

import (
	"strconv"

	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/tljson"
)

// Peer is a peer found in a decoded body.
type Peer struct {
	// Kind is wiretrace.PeerUser, PeerChat or PeerChannel.
	Kind string
	ID   int64
}

// String returns "user#123".
func (p Peer) String() string { return p.Kind + "#" + strconv.FormatInt(p.ID, 10) }

// peerFields maps constructors of peers to their kind and id field.
var peerFields = map[string][2]string{
	"inputPeerUser":            {wiretrace.PeerUser, "user_id"},
	"inputPeerUserFromMessage": {wiretrace.PeerUser, "user_id"},
	"inputUser":                {wiretrace.PeerUser, "user_id"},
	"peerUser":                 {wiretrace.PeerUser, "user_id"},
	"inputPeerChat":            {wiretrace.PeerChat, "chat_id"},
	"peerChat":                 {wiretrace.PeerChat, "chat_id"},
	"inputPeerChannel":         {wiretrace.PeerChannel, "channel_id"},
	"inputPeerChannelFromMessage": {
		wiretrace.PeerChannel, "channel_id",
	},
	"inputChannel": {wiretrace.PeerChannel, "channel_id"},
	"peerChannel":  {wiretrace.PeerChannel, "channel_id"},
}

func intValue(v any) (int64, bool) {
	switch v := v.(type) {
	case tljson.Long:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	}
	return 0, false
}

// PeersFrom collects the peers of a decoded TL value (tljson objects).
func PeersFrom(v any) []Peer {
	var out []Peer
	var walk func(v any)
	walk = func(v any) {
		switch v := v.(type) {
		case *tljson.Object:
			if f, ok := peerFields[v.Name()]; ok {
				if id, ok := intValue(v.Get(f[1])); ok {
					out = append(out, Peer{Kind: f[0], ID: id})
				}
			}
			for _, f := range v.Fields {
				walk(f.Value)
			}
		case []any:
			for _, e := range v {
				walk(e)
			}
		}
	}
	walk(v)
	return out
}

// JoinPeers maps the hashed ids of the trace (Result.Peers) to peers, using
// the run key (ui.key).
func (r *Result) JoinPeers(key wiretrace.IDKey, peers []Peer) {
	if len(key) == 0 {
		return
	}
	for _, p := range peers {
		r.Peers[key.Hash(wiretrace.PeerValue(p.Kind, p.ID))] = p.String()
	}
}

// JoinRawPeers maps raw ids (traces written with ui_raw_ids: the decimal
// PeerId.value) to peers.
func (r *Result) JoinRawPeers(peers []Peer) {
	for _, p := range peers {
		r.Peers[strconv.FormatUint(wiretrace.PeerValue(p.Kind, p.ID), 10)] = p.String()
	}
}
