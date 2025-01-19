package endpoints

import (
	"context"
	"net/http"

	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// deviceEndpoints registers the endpoints for the "Device" tagged endpoints.
func deviceEndpoints(api huma.API, session bluetooth.Session) {
	connectEndpoint(api, session)
	disconnectEndpoint(api, session)
	pairEndpoint(api, session)
	removeEndpoint(api, session)
	devicePropertiesEndpoint(api, session)
}

// devicePropertiesEndpoint registers the path "/device/{address}/properties".
func devicePropertiesEndpoint(api huma.API, session bluetooth.Session) {
	type DevicePropertiesOutput struct {
		Body bluetooth.DeviceData
	}

	huma.Register(api, huma.Operation{
		OperationID: "device-properties",
		Method:      http.MethodGet,
		Path:        "/device/{address}/properties",
		Summary:     "Properties",
		Description: "Fetches the properties of the device.",
		Tags:        []string{"Device"},
	}, func(_ context.Context, input *struct {
		AddressInput
	},
	) (*DevicePropertiesOutput, error) {
		deviceCall := session.Device(input.Address)

		properties, err := deviceCall.Properties()

		return &DevicePropertiesOutput{properties}, err
	})
}

// removeEndpoint registers the path "/device/{address}/remove".
func removeEndpoint(api huma.API, session bluetooth.Session) {
	huma.Register(api, huma.Operation{
		OperationID: "device-remove",
		Method:      http.MethodGet,
		Path:        "/device/{address}/remove",
		Summary:     "Remove",
		Description: "Removes a device from its associated adapter.",
		Tags:        []string{"Device"},
	}, func(_ context.Context, input *struct {
		AddressInput
	},
	) (*struct{}, error) {
		deviceCall := session.Device(input.Address)

		return nil, deviceCall.Remove()
	})
}

// pairEndpoint registers the path "/device/{address}/pair".
func pairEndpoint(api huma.API, session bluetooth.Session) {
	huma.Register(api, huma.Operation{
		OperationID: "device-pair",
		Method:      http.MethodGet,
		Path:        "/device/{address}/pair",
		Summary:     "Pairing",
		Description: "Starts a pairing process to an unpaired device in pairing mode. If the `cancel` parameter is specified, an ongoing pairing operation to the device, if it exists, will be stopped.",
		Tags:        []string{"Device"},
	}, func(_ context.Context, input *struct {
		AddressInput
		Cancel bool `doc:"Specifies if an ongoing pairing operation to the device should be cancelled." query:"cancel"`
	},
	) (*struct{}, error) {
		deviceCall := session.Device(input.Address)

		if input.Cancel {
			return nil, deviceCall.CancelPairing()
		}

		return nil, deviceCall.Pair()
	})
}

// connectEndpoint registers the path "/device/{address}/connect".
func connectEndpoint(api huma.API, session bluetooth.Session) {
	huma.Register(api, huma.Operation{
		OperationID: "device-connect",
		Method:      http.MethodGet,
		Path:        "/device/{address}/connect",
		Summary:     "Connection",
		Description: "Starts a connection process to a paired device. If a service profile UUID is specified, it will attempt to connect to it, otherwise a profile will be chosen and connected to automatically.",
		Tags:        []string{"Device"},
	}, func(_ context.Context, input *struct {
		AddressInput
		UUID uuid.UUID `doc:"The Bluetooth service profile UUID." example:"00001124-0000-1000-8000-00805f9b34fb" format:"uuid" query:"profile_uuid"`
	},
	) (*struct{}, error) {
		deviceCall := session.Device(input.Address)

		if input.UUID != uuid.Nil {
			return nil, deviceCall.ConnectProfile(input.UUID)
		}

		return nil, deviceCall.Connect()
	})
}

// disconnectEndpoint registers the path "/device/{address}/disconnect".
func disconnectEndpoint(api huma.API, session bluetooth.Session) {
	huma.Register(api, huma.Operation{
		OperationID: "device-disconnect",
		Method:      http.MethodGet,
		Path:        "/device/{address}/disconnect",
		Summary:     "Disconnection",
		Description: "Starts a disconnection process from a paired device. If a service profile UUID is specified, it will attempt to disconnect from it.",
		Tags:        []string{"Device"},
	}, func(_ context.Context, input *struct {
		AddressInput
		UUID uuid.UUID `doc:"The Bluetooth service profile UUID." example:"00001124-0000-1000-8000-00805f9b34fb" format:"uuid" query:"profile_uuid"`
	},
	) (*struct{}, error) {
		deviceCall := session.Device(input.Address)

		if input.UUID != uuid.Nil {
			return nil, deviceCall.DisconnectProfile(input.UUID)
		}

		return nil, deviceCall.Disconnect()
	})
}
