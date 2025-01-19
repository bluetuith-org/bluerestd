package endpoints

import (
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

// eventPublisher implements the eventbus.EventPublisher interface.
type eventPublisher struct {
	sender sse.Sender
}

// Publish sends the provided data to the registered event source.
func (e *eventPublisher) Publish(id uint, _ string, data any) {
	e.sender(sse.Message{
		ID:    int(id),
		Data:  data,
		Retry: 0,
	})
}

// AddressInput is used as the general input parameter for a Bluetooth address
// while registering paths that require it.
type AddressInput struct {
	Input   string `doc:"The Bluetooth MAC address." example:"11:22:33:AA:BB:CC" json:"address" path:"address"`
	Address bluetooth.MacAddress
}

// Resolve validates the input Bluetooth address.
func (a *AddressInput) Resolve(_ huma.Context) []error {
	mac, err := bluetooth.ParseMAC(a.Input)
	if err != nil {
		return []error{&huma.ErrorDetail{
			Message:  err.Error(),
			Location: "address",
			Value:    a.Input,
		}}
	}

	a.Address = mac

	return nil
}
