package endpoints

import (
	"context"
	"errors"
	"net/http"

	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/eventbus"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

// sessionEndpoints registers the endpoints for the "Session" tagged endpoints.
func sessionEndpoints(api huma.API, session bluetooth.Session) {
	eventsEndpoint(api)
	authEndpoint(api)

	adaptersEndpoint(api, session)
}

// adaptersEndpoint registers the path "/adapters".
func adaptersEndpoint(api huma.API, session bluetooth.Session) {
	type AdaptersOutput struct {
		Body []bluetooth.AdapterData
	}

	huma.Register(api, huma.Operation{
		OperationID: "adapters",
		Method:      http.MethodGet,
		Path:        "/adapters",
		Summary:     "Adapters",
		Tags:        []string{"Session"},
		Description: "Fetches all available adapters.",
	}, func(_ context.Context, _ *struct{}) (*AdaptersOutput, error) {
		return &AdaptersOutput{session.Adapters()}, nil
	})
}

// authEndpoint registers the path "/auth/{auth_id}/{reply}".
func authEndpoint(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "auth",
		Method:      http.MethodGet,
		Path:        "/auth/{auth_id}/{reply}",
		Summary:     "Authorization",
		Tags:        []string{"Session"},
		Description: "Enables responses to authorization requests, like device pairing or receiving file transfers.",
	}, func(_ context.Context, input *struct {
		Reply  string `doc:"The reply to an authorization request." enum:"yes,no" example:"yes" json:"reply,omitempty" path:"reply"`
		Reason string "doc:\"An optional user-specified reason if the reply is `no`.\" example:\"The user did not accept the request.\" json:\"reason,omitempty\" query:\"reason\""
		ID     int64  "doc:\"The authorization ID provided by the `auth` event.\" example:\"1\" path:\"auth_id\""
	},
	) (*struct{}, error) {
		if input.ID <= 0 {
			return nil, errors.New("invalid authorization ID")
		}

		ch, ok := requests.LoadAndDelete(input.ID)
		if !ok {
			return nil, errors.New("authorization ID not found")
		}

		ch <- authEventReply{input.Reason, input.Reply == "yes"}

		return nil, nil
	})
}

// eventsEndpoint registers the path "/events".
func eventsEndpoint(api huma.API) {
	newPublisher := func(sender sse.Sender) (*eventPublisher, eventbus.EventSubscriber) {
		eh := eventbus.NilHandler()

		return &eventPublisher{sender}, eh
	}

	sse.Register(api, huma.Operation{
		OperationID: "events",
		Method:      http.MethodGet,
		Path:        "/events",
		Tags:        []string{"Session"},
		Summary:     "Events",
		Description: "Subscribe to this EventSource for all Bluetooth events. For documentation on each watchable event, look at the *Responses* section.",
	}, map[string]any{
		"auth":         authRequestEvent{},
		"adapter":      bluetooth.AdapterEvent(),
		"error":        bluetooth.ErrorEvent(),
		"device":       bluetooth.DeviceEvent(),
		"mediaplayer":  bluetooth.MediaEvent(),
		"filetransfer": bluetooth.FileTransferEvent(),
	}, func(ctx context.Context, _ *struct{}, send sse.Sender) {
		eventbus.RegisterEventHandlers(newPublisher(send))
		defer eventbus.DisableEvents()

		<-ctx.Done()
	})
}
