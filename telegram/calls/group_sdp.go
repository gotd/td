package calls

import (
	"net"
	"strconv"
	"strings"
)

// groupJoinPayload is the JSON body sent in phone.joinGroupCall.params. It
// describes our ICE credentials, DTLS fingerprint and the SSRC of our outgoing
// audio, extracted from the local SDP offer (tgcalls GroupJoinPayload).
type groupJoinPayload struct {
	Ufrag        string             `json:"ufrag"`
	Pwd          string             `json:"pwd"`
	Fingerprints []groupFingerprint `json:"fingerprints"`
	Ssrc         int32              `json:"ssrc"`
}

type groupFingerprint struct {
	Hash        string `json:"hash"`
	Setup       string `json:"setup"`
	Fingerprint string `json:"fingerprint"`
}

// groupJoinResponse is the JSON returned in updateGroupCallConnection.params
// after joining: the SFU's transport parameters and media description (tgcalls
// GroupJoinResponsePayload).
type groupJoinResponse struct {
	Transport groupTransportDescription `json:"transport"`
	Audio     *groupMediaDescription    `json:"audio,omitempty"`
}

type groupTransportDescription struct {
	Ufrag        string             `json:"ufrag"`
	Pwd          string             `json:"pwd"`
	Fingerprints []groupFingerprint `json:"fingerprints"`
	Candidates   []groupCandidate   `json:"candidates"`
}

type groupCandidate struct {
	Foundation string `json:"foundation"`
	Component  string `json:"component"`
	Protocol   string `json:"protocol"`
	Priority   string `json:"priority"`
	IP         string `json:"ip"`
	Port       string `json:"port"`
	Type       string `json:"type"`
	Generation string `json:"generation"`
}

type groupMediaDescription struct {
	PayloadTypes  []groupPayloadType  `json:"payload-types"`
	RTPExtensions []groupRTPExtension `json:"rtp-hdrexts"`
}

type groupPayloadType struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	Clockrate     int             `json:"clockrate"`
	Channels      int             `json:"channels,omitempty"`
	FeedbackTypes []groupFeedback `json:"rtcp-fbs,omitempty"`
}

type groupFeedback struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
}

type groupRTPExtension struct {
	ID  int    `json:"id"`
	URI string `json:"uri"`
}

// extractSDPParams pulls the ICE ufrag/pwd and DTLS fingerprint out of a local
// SDP description.
func extractSDPParams(sdp string) (ufrag, pwd, fingerprint, hash string) {
	for _, line := range strings.Split(sdp, "\r\n") {
		switch {
		case strings.HasPrefix(line, "a=ice-ufrag:"):
			ufrag = strings.TrimPrefix(line, "a=ice-ufrag:")
		case strings.HasPrefix(line, "a=ice-pwd:"):
			pwd = strings.TrimPrefix(line, "a=ice-pwd:")
		case strings.HasPrefix(line, "a=fingerprint:"):
			if parts := strings.SplitN(strings.TrimPrefix(line, "a=fingerprint:"), " ", 2); len(parts) == 2 {
				hash, fingerprint = parts[0], parts[1]
			}
		}
	}
	return ufrag, pwd, fingerprint, hash
}

// buildAnswerSDP renders an SDP answer for the SFU from its JSON response. The
// SFU is ICE-lite and acts as the DTLS client (a=setup:active), so we are the
// passive DTLS server.
func buildAnswerSDP(resp groupJoinResponse) string {
	t := resp.Transport
	port := remotePort(t.Candidates)
	conn := remoteConnLine(t.Candidates)

	payloads := []string{}
	if resp.Audio != nil {
		for _, pt := range resp.Audio.PayloadTypes {
			payloads = append(payloads, strconv.Itoa(pt.ID))
		}
	}
	if len(payloads) == 0 {
		payloads = []string{"111"}
	}

	lines := []string{
		"v=0",
		"o=- 1 2 IN IP4 0.0.0.0",
		"s=-",
		"t=0 0",
		"a=group:BUNDLE 0",
		"a=ice-lite",
		"m=audio " + strconv.Itoa(port) + " RTP/SAVPF " + strings.Join(payloads, " "),
		conn,
		"a=mid:0",
		"a=ice-ufrag:" + t.Ufrag,
		"a=ice-pwd:" + t.Pwd,
	}
	if len(t.Fingerprints) > 0 {
		lines = append(lines, "a=fingerprint:"+t.Fingerprints[0].Hash+" "+t.Fingerprints[0].Fingerprint)
	}
	lines = append(lines, "a=setup:active")

	for _, c := range t.Candidates {
		if net.ParseIP(c.IP) == nil {
			continue
		}
		lines = append(lines, strings.Join([]string{
			"a=candidate:" + c.Foundation, c.Component, c.Protocol, c.Priority,
			c.IP, c.Port, "typ", c.Type, "generation", c.Generation,
		}, " "))
	}

	if resp.Audio != nil {
		for _, pt := range resp.Audio.PayloadTypes {
			rtpmap := "a=rtpmap:" + strconv.Itoa(pt.ID) + " " + pt.Name + "/" + strconv.Itoa(pt.Clockrate)
			if pt.Channels > 1 {
				rtpmap += "/" + strconv.Itoa(pt.Channels)
			}
			lines = append(lines, rtpmap)
			if pt.Name == "opus" {
				lines = append(lines, "a=fmtp:"+strconv.Itoa(pt.ID)+" minptime=10;useinbandfec=1")
			}
			for _, fb := range pt.FeedbackTypes {
				fbLine := "a=rtcp-fb:" + strconv.Itoa(pt.ID) + " " + fb.Type
				if fb.Subtype != "" {
					fbLine += " " + fb.Subtype
				}
				lines = append(lines, fbLine)
			}
		}
		seen := map[int]bool{}
		for _, ext := range resp.Audio.RTPExtensions {
			if seen[ext.ID] {
				continue
			}
			seen[ext.ID] = true
			lines = append(lines, "a=extmap:"+strconv.Itoa(ext.ID)+" "+ext.URI)
		}
	}

	lines = append(lines, "a=rtcp-mux", "a=sendrecv", "")
	return strings.Join(lines, "\r\n")
}

func remoteIP(candidates []groupCandidate) string {
	for _, c := range candidates {
		if net.ParseIP(c.IP) != nil {
			return c.IP
		}
	}
	return "0.0.0.0"
}

func remoteConnLine(candidates []groupCandidate) string {
	ip := net.ParseIP(remoteIP(candidates))
	if ip != nil && ip.To4() == nil {
		return "c=IN IP6 " + ip.String()
	}
	return "c=IN IP4 " + remoteIP(candidates)
}

func remotePort(candidates []groupCandidate) int {
	for _, c := range candidates {
		if net.ParseIP(c.IP) != nil {
			if p, err := strconv.Atoi(c.Port); err == nil {
				return p
			}
		}
	}
	return 1
}
