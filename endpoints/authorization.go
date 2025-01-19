package endpoints

import (
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/eventbus"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"
)

// authRequestEvent describes a request authorization event.
type authRequestEvent struct {
	PairingParams  *authPairingEvent  "doc:\"The parameters of the `pairing` authorization request.\" json:\"pairing_params,omitempty\""
	TransferParams *authTransferEvent "doc:\"The parameters of the `transfer` authorization request.\" json:\"transfer_params,omitempty\""
	AuthType       string             "doc:\"The type of the authorization request.\" enum:\"pairing,transfer\" json:\"auth_type,omitempty\""

	ID            int64 "doc:\"The ID of the authorization request.\" json:\"auth_id,omitempty\""
	ReplyRequired bool  "doc:\"If this parameter is set to 'true', use the `/auth/{auth_id}/{reply}` endpoint to respond to this request, otherwise ignore.\" json:\"reply_required,omitempty\""
}

// authPairingEvent describes a pairing authorization event.
type authPairingEvent struct {
	ServiceUUID *uuid.UUID `doc:"The service profile UUID."   json:"uuid,omitempty"`
	PairingType string     `doc:"The type of the pairing authorization request." enum:"display-pincode,display-passkey,confirm-passkey,authorize-pairing,authorize-service" json:"pairing_type,omitempty"`

	Pincode string `doc:"The provided pincode value." json:"pincode,omitempty"`
	Passkey uint32 `doc:"The provided passkey value." json:"passkey,omitempty"`
	Entered uint16 `doc:"The entered passkey value."  json:"entered,omitempty"`

	Address bluetooth.MacAddress `doc:"The address of the device."  json:"address,omitempty"`
}

// authTransferEvent describes a transfer authorization event.
type authTransferEvent struct {
	Path           string                     `doc:"The path to the file."       json:"path,omitempty"`
	FileProperties bluetooth.FileTransferData `doc:"The properties of the file." json:"file_properties,omitempty"`
}

// authEventReply describes a reply to an authorization event.
type authEventReply struct {
	reason string
	reply  bool
}

// authEventID is the authorization event ID.
type authEventID uint

// authEvent is the defined authorization event ID.
const authEvent = authEventID(100)

// requests store the pending authorization requests.
var requests = xsync.NewMapOf[int64, chan authEventReply]()

// Authorizer implements the bluetooth.SessionAuthorizer interface.
type Authorizer struct {
	id *xsync.Counter
}

// NewAuthorizer returns a new authorizer to use as the session's authorization handler.
func NewAuthorizer() *Authorizer {
	return &Authorizer{id: xsync.NewCounter()}
}

// AuthorizeTransfer sends a "transfer" authentication request.
func (a *Authorizer) AuthorizeTransfer(timeout bluetooth.AuthTimeout, path string, props bluetooth.FileTransferData) error {
	return a.sendAndWait(timeout, authRequestEvent{
		AuthType:      "transfer",
		ReplyRequired: true,
		TransferParams: &authTransferEvent{
			Path:           path,
			FileProperties: props,
		},
	})
}

// DisplayPinCode sends a "display-pincode" pairing authentication request.
func (a *Authorizer) DisplayPinCode(_ bluetooth.AuthTimeout, address bluetooth.MacAddress, pincode string) error {
	a.send(authRequestEvent{
		AuthType:      "pairing",
		ReplyRequired: false,
		PairingParams: &authPairingEvent{
			PairingType: "display-pincode",
			Address:     address,
			Pincode:     pincode,
		},
	})

	return nil
}

// DisplayPasskey sends a "display-passkey" pairing authentication request.
func (a *Authorizer) DisplayPasskey(_ bluetooth.AuthTimeout, address bluetooth.MacAddress, passkey uint32, entered uint16) error {
	a.send(authRequestEvent{
		AuthType:      "pairing",
		ReplyRequired: false,
		PairingParams: &authPairingEvent{
			PairingType: "display-passkey",
			Address:     address,
			Passkey:     passkey,
			Entered:     entered,
		},
	})

	return nil
}

// ConfirmPasskey sends a "confirm-passkey" pairing authentication request.
func (a *Authorizer) ConfirmPasskey(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, passkey uint32) error {
	return a.sendAndWait(timeout, authRequestEvent{
		AuthType:      "pairing",
		ReplyRequired: true,
		PairingParams: &authPairingEvent{
			PairingType: "confirm-passkey",
			Address:     address,
			Passkey:     passkey,
		},
	})
}

// AuthorizePairing sends a "authorize-pairing" pairing authentication request.
func (a *Authorizer) AuthorizePairing(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress) error {
	return a.sendAndWait(timeout, authRequestEvent{
		AuthType:      "pairing",
		ReplyRequired: true,
		PairingParams: &authPairingEvent{
			PairingType: "authorize-pairing",
			Address:     address,
		},
	})
}

// AuthorizeService sends a "authorize-service" pairing authentication request.
func (a *Authorizer) AuthorizeService(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, uuid uuid.UUID) error {
	return a.sendAndWait(timeout, authRequestEvent{
		AuthType:      "pairing",
		ReplyRequired: true,
		PairingParams: &authPairingEvent{
			PairingType: "authorize-service",
			Address:     address,
			ServiceUUID: &uuid,
		},
	})
}

// send publishes the authorization request to the event stream.
func (a *Authorizer) send(data authRequestEvent) int64 {
	a.id.Inc()
	data.ID = a.id.Value()

	eventbus.Publish(authEvent, data)

	return data.ID
}

// sendAndWait publishes the authorization request to the event stream and waits for a response.
func (a *Authorizer) sendAndWait(timeout bluetooth.AuthTimeout, data authRequestEvent) error {
	var reply authEventReply

	ch := make(chan authEventReply, 1)
	requests.Store(a.send(data), ch)
	select {
	case <-timeout.Done():
	case reply = <-ch:
	}

	if reply.reply {
		return nil
	}

	return reply
}

func (i authEventID) String() string {
	return "auth"
}

func (i authEventID) Value() uint {
	return uint(i)
}

func (a authEventReply) Error() string {
	if a.reply {
		return ""
	}

	reason := "The authorization request was not accepted."
	if a.reason != "" {
		reason = a.reason
	}

	return reason
}
