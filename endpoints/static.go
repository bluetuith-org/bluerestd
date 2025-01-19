package endpoints

import (
	_ "embed"

	"github.com/danielgtaylor/huma/v2"
)

//go:embed assets/scalar.js
var staticDocumentJS string

var staticHTML = `<!doctype html>
<html>
  <head>
    <title>Scalar API Reference</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.yaml"></script>
    <script>` + staticDocumentJS + `</script>
  </body>
</html>`

var staticAPIInfo = &huma.Info{
	Title:       "Bluetooth API",
	Description: staticPageDescription,
	Contact: &huma.Contact{
		Name: "bluetuith-org",
		URL:  "https://github.com/bluetuith-org",
	},
	License: &huma.License{
		Name:       "MIT",
		Identifier: "MIT",
		URL:        "https://github.com/bluetuith-org/bluerestd/blob/master/LICENSE",
	},
	Version: "0.0.1",
}

var staticPageDescription = `
This documentation describes the complete OpenAPI specification of this instance.

To begin, navigate to the [Session](#tag/session) section.
`

var staticTagDescriptions = map[string]string{
	"Session": `
The session is mostly event-driven, and events must be subscribed to
for receiving updates about the various components in the session.
This is especially required for authenticating any pairing/file transfer requests.

- Subscribe to the EventSource using the [Events endpoint](#tag/session/GET/events).
- For authorization requests, watch the *"auth"* event. All *"auth"* events return
  an authorization ID (auth_id), which can be used with the [Authorization endpoint](#tag/session/GET/auth/{auth_id}/{reply}). 
- Then, to fetch a list of available adapters, use the [Adapters endpoint](#tag/session/GET/adapters).

To interact with an adapter from the list, go to the [Adapter](#tag/adapter) section.
`,

	"Adapter": `
These set of endpoints interact with an individual adapter.
The **address** parameter refers to the **adapter's** Bluetooth address, and is required. 

- To change each state/mode of the adapter (powered, pairable, discoverable and discovery), use the
  [States endpoint](#tag/adapter/GET/adapter/{address}/states).
- To view the properties of the adapter, use the [Properties endpoint](#tag/adapter/GET/adapter/{address}/properties).
- To fetch a list of devices associated with this adapter, use the [Devices endpoint](#tag/adapter/GET/adapter/{address}/devices).

To interact with a device from the list, go to the [Device](#tag/device) section.

#### Note
When discovery is enabled, all discovered devices will be published to the **/event** stream, with the *event_name* as *'device'*, and with *event_action* as *'added'*.
`,

	"Device": `
These set of endpoints interact with an individual device.
The **address** parameter refers to the **device's** Bluetooth address, and is required. 

## Pairing
The device pairing process usually follows after a device discovery is in progress on the adapter.

Watch the following events:
- The *auth* event, filter for the *pairing* 'auth_type' to get pairing authorization requests.
- The *device* event, filter for the *added* 'event_action' to get discovered devices.

To initiate pairing on a discovered device using its Bluetooth address:
- Use the [Pairing endpoint](#tag/device/GET/device/{address}/pair) to start the pairing process.
- Then, an authorization request will be posted to the **/event** stream as an *auth* event
  Note the 'auth_id' of the event.
- Using the [Authorization endpoint](#tag/session/GET/auth/{auth_id}/{reply}), respond to the authorization request using the noted 'auth_id'.
- If the authorization request was confirmed, the device will be paired to the adapter.

## Other functions
The following endpoints function only on paired devices.

Use the:
- [Connect endpoint](#tag/device/GET/device/{address}/connect) and [Disconnect endpoint](#tag/device/GET/device/{address}/disconnect) to connect to/disconnect from a device.
- [Properties endpoint](#tag/device/GET/device/{address}/properties) to view the properties of a device.
- [Remove endpoint](#tag/device/GET/device/{address}/remove) to remove a device from its associated adapter. 
`,

	"Network": `
These set of endpoints extend the core functionalities of the [device endpoints](#tag/device),
and provide methods to tether to/untether from a device's internet connection using the PANU
or DUN Bluetooth profiles (BNEP).

The **address** parameter refers to the **device's** Bluetooth address, and is required. 

Use the: 
- [Connection endpoint](#tag/network/GET/device/{address}/network_connect/{connection_type}) to start the tethering process.
- [Disconnection endpoint](#tag/network/GET/device/{address}/network_disconnect) to untether from an existing tethered connection to the device. 
`,

	"File Transfer": `
These set of endpoints extend the core functionalities of the [device endpoints](#tag/device),
and provide methods to transfer files to a device using the Bluetooth OBEX Object Push profile.

The **address** parameter refers to the **device's** Bluetooth address, and is required. 

# Receiving files
Watch the following events:
- The *auth* event, filter for the *transfer* 'auth_type' to get file transfer authorization requests.

On receiving the authorization request, respond to the authorization request with the noted 'auth_id'
using the [Authorization endpoint](#tag/session/GET/auth/{auth_id}/{reply}).
If the authorization request was confirmed, the file transfer will start to be received.

# Sending files
Start by submitting a list of files to be transferred to the [Start Transfers endpoint](#tag/file-transfer/POST/device/{address}/start_file_transfer).

Then, watch the following events:
- The *filetransfer* event, filter for the *updated* 'event_action' to get the status of the file transfers.

To cancel an ongoing transfer, use the [Stop Transfers endpoint](#tag/file-transfer/GET/device/{address}/stop_file_transfer).
`,
}
