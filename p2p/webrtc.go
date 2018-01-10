package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	webrtc "github.com/keroserene/go-webrtc"
)

/* WebRTC 1.0 Protocol

   See https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Connectivity
   for a description of the specific connectivity steps we execute here.
   The steps, numbered 1 to 10, are refered to in code comments below.

   For the full protocol spec, see: https://www.w3.org/TR/webrtc/
*/

const (
	// TODO: remove when Orchid nodes implement STUN
	stunServer = "stun:stun.l.google.com:19302"
)

type WebRTCPeer struct {
	RefURL   *url.URL
	PC       *webrtc.PeerConnection
	DCs      []*webrtc.DataChannel
	IceCands []*webrtc.IceCandidate
}

// TODO: this JSON schema is temp in lieu of first protocol spec lockdown
type SDPAndIce struct {
	Description webrtc.SessionDescription `json:"description"`
	Candidates  []*webrtc.IceCandidate
}

type Offer struct {
	Inner SDPAndIce `json:"offer"`
}

type Answer struct {
	Inner SDPAndIce `json:"answer"`
}

// see orchid-core/src/index.ts interface BackResponse
type BackResponse struct {
	Pub         string `json:"nodePub"`
	ETHBlock    uint32 `json:"ethBlock"`
	PoWSolution uint64 `json:"powSolution"` // TODO: Equihash
	Answer      string `json:"answerSDP"`
}

func NewWebRTCPeer(ref *url.URL) (*WebRTCPeer, error) {
	// Prior to step 1:
	// configure go-webrtc lib, create a new PeerConnection and add
	// event listeners for Ice, signaling and connection events.
	webrtc.SetLoggingVerbosity(1)
	config := webrtc.NewConfiguration(
		webrtc.OptionIceServer(stunServer),
	)

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Error("webrtc.NewPeerConnection", "err", err)
		return nil, err
	}

	cands := []*webrtc.IceCandidate{}
	candChan := make(chan webrtc.IceCandidate, 2)

	// ICE Events
	pc.OnIceCandidate = func(c webrtc.IceCandidate) {
		log.Debug("OnIceCandidate: ", "cand", c)
		candChan <- c
	}
	pc.OnIceCandidateError = func() {
		log.Debug("OnIceCandidateError: ")
	}
	pc.OnIceConnectionStateChange = func(webrtc.IceConnectionState) {
		log.Debug("OnIceConnectionStateChange: ")
	}
	pc.OnIceGatheringStateChange = func(webrtc.IceGatheringState) {
		log.Debug("OnIceGatheringStateChange: ")
	}
	pc.OnIceComplete = func() {
		log.Debug("OnIceComplete: ")
		close(candChan)
	}

	// Other PeerConnection Events
	pc.OnSignalingStateChange = func(s webrtc.SignalingState) {
		log.Debug("OnSignalingStateChange: ", "state", s)
	}
	pc.OnConnectionStateChange = func(webrtc.PeerConnectionState) {
		log.Debug("OnConnectionStateChange: ")
	}

	// To trigger ICE, we have to create a RTCDataChannel before
	// we create the signaling offer
	dc, err := pc.CreateDataChannel("0")
	if err != nil {
		log.Error("CreateDataChannel", "err", err)
		return nil, err
	}

	// Step 1:
	offerSDP, err := pc.CreateOffer()
	if err != nil {
		log.Error("CreateOffer", "err", err)
		return nil, err
	}

	// Step 2:
	err = pc.SetLocalDescription(offerSDP)
	if err != nil {
		log.Error("SetLocalDescription", "err", err)
		return nil, err
	}

	// Block on ice candidates
	for cand := range candChan {
		cands = append(cands, &cand)
	}

	// Step 3: transmit WebRTC offer and ICE candidates over signaling channel;
	//         HTTP(S) for now
	response := Offer{SDPAndIce{*offerSDP, cands}}
	b, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)

	// This triggers step 4-8 at the remote
	resp, err := http.Post("http://"+ref.Host, "application/json", buf)
	if err != nil {
		return nil, err
	}
	// Step 9: receive the answer (validate response)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("WebRTC signaling over HTTP failed, resp.StatusCode: %d\n", resp.StatusCode)
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Could not read HTTP response body", "err", err)
		return nil, err
	}

	answer := new(Answer)
	err = json.Unmarshal(b, &answer)
	sdpAndIce := answer.Inner
	if err != nil {
		log.Error("Could not decode HTTP response body JSON", "err", err)
		return nil, err
	}
	answerSDP := sdpAndIce.Description

	// Step 10: (validates the received SDP)
	err = pc.SetRemoteDescription(&answerSDP)
	if err != nil {
		log.Error("SetRemoteDescription", "err", err)
		return nil, err
	}

	// Add candidates from peer
	for _, c := range sdpAndIce.Candidates {
		if c == nil || c.Candidate == "" {
			continue // TODO: verify if correct behaviour
		}
		err = pc.AddIceCandidate(*c)
		if err != nil {
			log.Error("AddIceCandidate", "err", err)
			return nil, err
		}
	}

	peer := WebRTCPeer{
		ref,
		pc,
		[]*webrtc.DataChannel{dc},
		cands,
	}

	return &peer, nil
}

func NewExit(b []byte, dcReady chan *webrtc.DataChannel) ([]byte, *WebRTCPeer, error) {
	offer := new(Offer)
	err := json.Unmarshal(b, offer)
	if err != nil {
		log.Error("Parsing WebRTC Offer", "err", err)
		return nil, nil, err
	}
	log.Debug("offer", "struct", offer)
	sdpAndIce := offer.Inner
	//log.Debug("WebRTC Offer", "type", sdpAndIce.Description.Type, "sdp", sdpAndIce.Description.Sdp)
	for i, c := range sdpAndIce.Candidates {
		if c != nil {
			log.Debug("RECEIVED ICE", "index", i, "candidate", c.Candidate, "sdpMid", c.SdpMid, "SdpMLineIndex", c.SdpMLineIndex)
		} else {
			log.Warn("Received nil ICE Candidate")
		}
	}

	// At this point we have what looks like a valid WebRTC offer SDP,
	// with steps 1,2,3 done by the caller and we execute step 4:
	webrtc.SetLoggingVerbosity(3)
	config := webrtc.NewConfiguration(
		webrtc.OptionIceServer("stun:stun.l.google.com:19302"),
	)
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Error("webrtc.NewPeerConnection", "err", err)
		return nil, nil, err
	}

	// Listen to our own candidates
	cands := make([]*webrtc.IceCandidate, 2)
	candChan := make(chan webrtc.IceCandidate, 2)
	// ICE Events
	pc.OnIceCandidate = func(c webrtc.IceCandidate) {
		log.Debug("OnIceCandidate: ", "cand", c)
		candChan <- c
	}
	pc.OnIceCandidateError = func() {
		log.Debug("OnIceCandidateError: ")
	}
	pc.OnIceConnectionStateChange = func(webrtc.IceConnectionState) {
		log.Debug("OnIceConnectionStateChange: ")
	}
	pc.OnIceGatheringStateChange = func(webrtc.IceGatheringState) {
		log.Debug("OnIceGatheringStateChange: ")
	}
	pc.OnIceComplete = func() {
		log.Debug("OnIceComplete: ")
		close(candChan)
	}
	// Other PeerConnection Events
	pc.OnSignalingStateChange = func(s webrtc.SignalingState) {
		log.Debug("OnSignalingStateChange: ", "state", s)
	}
	pc.OnConnectionStateChange = func(webrtc.PeerConnectionState) {
		log.Debug("OnConnectionStateChange: ")
	}

	var dc *webrtc.DataChannel
	pc.OnDataChannel = func(d *webrtc.DataChannel) {
		log.Debug("OnDataChannel: ")
		d.OnMessage = func(msg []byte) {
			log.Debug("temp dc.OnMessage", "msg", msg)
		}
		dcReady <- d
	}

	err = pc.SetRemoteDescription(&sdpAndIce.Description)
	if err != nil {
		log.Error("SetRemoteDescription", "err", err)
		return nil, nil, err
	}

	// Add candidates from peer
	for _, c := range sdpAndIce.Candidates {
		if c == nil || c.Candidate == "" {
			continue // TODO: verify if should we skip here
		}
		log.Debug("ICE", "adding", c, "c.candidate", c.Candidate)
		err = pc.AddIceCandidate(*c)
		if err != nil {
			log.Error("AddIceCandidate", "err", err)
			return nil, nil, err
		}
	}

	// Step 5: TODO: anything else we need locally, e.g. resource
	//               allocations (out of scope of the webrtc spec)

	// Step 6:
	answerSDP, err := pc.CreateAnswer()
	if err != nil {
		log.Error("CreateAnswer", "err", err)
		return nil, nil, err
	}

	// Step 7:
	err = pc.SetLocalDescription(answerSDP)
	if err != nil {
		log.Error("SetLocalDescription", "err", err)
		return nil, nil, err
	}

	// Block on ice candidates
	for cand := range candChan {
		cands = append(cands, &cand)
	}

	// Step 8:
	// TODO: for now we send back Orchid specific fields alongside
	//       the answer SDP. For live network everything must be encrypted
	resp := Answer{SDPAndIce{*answerSDP, cands}}
	respBuf, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, err
	}
	//log.Info("response", "bytes", string(respBuf))

	peer := WebRTCPeer{
		nil,
		pc,
		[]*webrtc.DataChannel{dc},
		cands,
	}

	return respBuf, &peer, nil
}

/* DCReadWriteCloser wraps webrtc.DataChannel with a mutex for
   concurrent access and a read buffer and closed flag to implement
   the io.ReadWriterCloser interface as a more generic way of interfacing
   with the TCPProxy
*/
type DCReadWriteCloser struct {
	Mutex   sync.Mutex
	DC      *webrtc.DataChannel
	ReadBuf []byte
	Closed  bool
}

func NewDCReadWriteCloser(dc *webrtc.DataChannel) *DCReadWriteCloser {
	d := &DCReadWriteCloser{
		sync.Mutex{},
		dc,
		make([]byte, transferBufSize),
		false,
	}

	dc.OnMessage = func(msg []byte) {
		log.Debug("DCReadWriteCloser dc.OnMessage 0")
		d.Mutex.Lock()
		d.ReadBuf = append(d.ReadBuf, msg...)
		log.Debug("DCReadWriteCloser dc.OnMessage 1", "readBuf", d.ReadBuf)
		d.Mutex.Unlock()
	}
	dc.OnClose = func() {
		d.Mutex.Lock()
		d.Closed = true
		d.Mutex.Unlock()
	}
	return d
}

func (d *DCReadWriteCloser) Read(p []byte) (n int, err error) {
	log.Debug("DCReadWriteCloser Read 0")
	defer d.Mutex.Unlock()
	d.Mutex.Lock()

	if d.Closed {
		return 0, io.EOF
	}
	n = copy(p, d.ReadBuf)
	log.Debug("DCReadWriteCloser Read", "p", p)
	return n, nil
}

func (d *DCReadWriteCloser) Write(p []byte) (n int, err error) {
	log.Debug("DCReadWriteCloser Write 0", "p", p)
	defer d.Mutex.Unlock()
	d.Mutex.Lock()
	// copy to new slice since webrtc.DataChannel accesses the
	// passed byte slice using cgo & unsafe pointers and Writer interface
	// implementations must not retain p
	c := make([]byte, len(p))
	copy(c, p)
	d.DC.Send(c)
	log.Debug("DCReadWriteCloser Write 1", "p", p)
	return len(p), nil
}

func (d *DCReadWriteCloser) Close() (err error) {
	log.Debug("DCReadWriteCloser Close")
	defer d.Mutex.Unlock()
	d.Mutex.Lock()

	d.Closed = true
	err = d.DC.Close()
	return
}
