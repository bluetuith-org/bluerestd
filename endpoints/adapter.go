package endpoints

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/danielgtaylor/huma/v2"
)

// adapterEndpoints registers the endpoints for the "Adapter" tagged endpoints.
func adapterEndpoints(api huma.API, session bluetooth.Session) {
	devicesEndpoint(api, session)
	statesEndpoint(api, session)
	adapterPropertiesEndpoint(api, session)
}

// devicesEndpoint registers the path "/adapter/{address}/devices".
func devicesEndpoint(api huma.API, session bluetooth.Session) {
	type AdapterDevicesOutput struct {
		Body []bluetooth.DeviceData
	}

	huma.Register(api, huma.Operation{
		OperationID: "adapter-devices",
		Method:      http.MethodGet,
		Path:        "/adapter/{address}/devices",
		Summary:     "Devices",
		Description: "Fetches the devices associated with an adapter.",
		Tags:        []string{"Adapter"},
	}, func(_ context.Context, input *struct {
		AddressInput
	},
	) (*AdapterDevicesOutput, error) {
		adapterCall := session.Adapter(input.Address)

		devices, err := adapterCall.Devices()

		return &AdapterDevicesOutput{devices}, err
	})
}

// adapterPropertiesEndpoint registers the path "/adapter/{address}/properties".
func adapterPropertiesEndpoint(api huma.API, session bluetooth.Session) {
	type AdapterPropertiesOutput struct {
		Body bluetooth.AdapterData
	}

	huma.Register(api, huma.Operation{
		OperationID: "adapter-properties",
		Method:      http.MethodGet,
		Path:        "/adapter/{address}/properties",
		Summary:     "Properties",
		Description: "Fetches the properties of an adapter.",
		Tags:        []string{"Adapter"},
	}, func(_ context.Context, input *struct {
		AddressInput
	},
	) (*AdapterPropertiesOutput, error) {
		adapterCall := session.Adapter(input.Address)

		properties, perr := adapterCall.Properties()
		if perr != nil {
			return nil, perr
		}

		return &AdapterPropertiesOutput{properties}, nil
	})
}

// statesEndpoint registers the path "/adapter/{address}/states".
func statesEndpoint(api huma.API, session bluetooth.Session) {
	type AdapterStatesInput struct {
		Powered      string `doc:"Set the adapter's powered state."            enum:"enable,disable" query:"powered"`
		Pairable     string `doc:"Set the adapter's pairable state."           enum:"enable,disable" query:"pairable"`
		Discoverable string `doc:"Set the adapter's discoverable state."       enum:"enable,disable" query:"discoverable"`
		Discovery    string `doc:"Toggle the adapter's device discovery mode." enum:"enable,disable" query:"discovery"`
	}

	type AdapterStatesOutput struct {
		Body struct {
			PoweredState      string `doc:"The adapter's powered state."               enum:"enabled,disabled" json:"powered,omitempty"`
			PairableState     string `doc:"The adapter's pairable state."              enum:"enabled,disabled" json:"pairable,omitempty"`
			DiscoverableState string `doc:"The adapter's discoverable state."          enum:"enabled,disabled" json:"discoverable,omitempty"`
			DiscoveryState    string `doc:"The adapter's device discovery mode state." enum:"enabled,disabled" json:"discovery,omitempty"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "adapter-states",
		Method:      http.MethodGet,
		Path:        "/adapter/{address}/states",
		Summary:     "States",
		Description: "This endpoint, when called by itself, fetches the different states (powered, pairable, discoverable and device discovery) of an adapter. Use the **query parameters** to `enable` or `disable` each state. Note that when **discovery** is **enabled**, all discovered devices will be published to the `/event` stream, with the ***event-name*** as *'device'*, and with ***event-action*** as *'added'*.",
		Tags:        []string{"Adapter"},
	}, func(_ context.Context, input *struct {
		AdapterStatesInput
		AddressInput
	},
	) (*AdapterStatesOutput, error) {
		states := &AdapterStatesOutput{}
		adapterCall := session.Adapter(input.Address)

		inputs := []struct {
			EnableFunc        func() error
			DisableFunc       func() error
			SetStatesProperty func(string)
			InputToCheck      string
		}{
			{
				InputToCheck: input.Discovery,
				EnableFunc:   adapterCall.StartDiscovery,
				DisableFunc:  adapterCall.StopDiscovery,
				SetStatesProperty: func(toggle string) {
					states.Body.DiscoveryState = toggle
				},
			},
			{
				InputToCheck: input.Discoverable,
				EnableFunc: func() error {
					return adapterCall.SetDiscoverableState(true)
				},
				DisableFunc: func() error {
					return adapterCall.SetDiscoverableState(false)
				},
				SetStatesProperty: func(toggle string) {
					states.Body.DiscoverableState = toggle
				},
			},
			{
				InputToCheck: input.Pairable,
				EnableFunc: func() error {
					return adapterCall.SetPairableState(true)
				},
				DisableFunc: func() error {
					return adapterCall.SetPairableState(false)
				},
				SetStatesProperty: func(toggle string) {
					states.Body.PairableState = toggle
				},
			},
			{
				InputToCheck: input.Powered,
				EnableFunc: func() error {
					return adapterCall.SetPoweredState(true)
				},
				DisableFunc: func() error {
					return adapterCall.SetPoweredState(false)
				},
				SetStatesProperty: func(toggle string) {
					states.Body.PoweredState = toggle
				},
			},
		}

		var errs error

		var emptyInputs int

		for _, in := range inputs {
			var err error

			var state string

			switch in.InputToCheck {
			case "enable":
				err = in.EnableFunc()
				state = "enabled"
			case "disable":
				err = in.DisableFunc()
				state = "disabled"
			case "":
				emptyInputs++

				continue
			}

			if err != nil {
				if errs == nil {
					errs = err
				} else {
					errs = fmt.Errorf("%w, %w", errs, err)
				}
			}

			in.SetStatesProperty(state)
		}

		if errs != nil {
			return nil, errs
		}

		if emptyInputs == len(inputs) {
			properties, perr := adapterCall.Properties()
			if perr != nil {
				return nil, perr
			}

			states.Body.DiscoverableState = toggleStr(properties.Discoverable)
			states.Body.PairableState = toggleStr(properties.Pairable)
			states.Body.DiscoveryState = toggleStr(properties.Discovering)
			states.Body.PoweredState = toggleStr(properties.Powered)
		}

		return states, nil
	})
}

// toggleStr returns "enabled" or "disabled" depending on the provided value.
func toggleStr(val bool) string {
	if val {
		return "enabled"
	}

	return "disabled"
}
