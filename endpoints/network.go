package endpoints

import (
	"context"
	"net/http"
	"strings"

	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/danielgtaylor/huma/v2"
)

// networkEndpoints registers the endpoints for the "Network" tagged endpoints.
func networkEndpoints(api huma.API, session bluetooth.Session) {
	connectNetworkEndpoint(api, session)
	disconnectNetworkEndpoint(api, session)
}

// connectNetworkEndpoint registers the path "/device/{address}/network_connect/{connection_type}".
func connectNetworkEndpoint(api huma.API, session bluetooth.Session) {
	type NetworkTypeInput struct {
		Type bluetooth.NetworkType `default:"panu" doc:"The type of Bluetooth profile to use to tether to the device's internet connection." enum:"panu,dun" json:"connection_type" path:"connection_type"`
	}

	huma.Register(api, huma.Operation{
		OperationID: "device-network-connect",
		Method:      http.MethodGet,
		Path:        "/device/{address}/network_connect/{connection_type}",
		Summary:     "Connection (PANU, DUN)",
		Description: "Attempts to tether to the internet connection of the device.",
		Tags:        []string{"Network"},
	}, func(_ context.Context, input *struct {
		AddressInput
		NetworkTypeInput
	},
	) (*struct{}, error) {
		device, err := session.Device(input.Address).Properties()
		if err != nil {
			return nil, err
		}

		networkName := device.Name + " Connection (" + device.Address.String() + ", " + strings.ToUpper(input.Type.String()) + ")"

		return nil, session.Network(input.Address).Connect(networkName, input.Type)
	})
}

// disconnectNetworkEndpoint registers the path "/device/{address}/network_disconnect".
func disconnectNetworkEndpoint(api huma.API, session bluetooth.Session) {
	huma.Register(api, huma.Operation{
		OperationID: "device-network-disconnect",
		Method:      http.MethodGet,
		Path:        "/device/{address}/network_disconnect",
		Summary:     "Disconnection",
		Description: "Attempts to untether from the internet connection of the device.",
		Tags:        []string{"Network"},
	}, func(_ context.Context, input *struct {
		AddressInput
	},
	) (*struct{}, error) {
		return nil, session.Network(input.Address).Disconnect()
	})
}
