package endpoints

import (
	"context"
	"net/http"

	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/danielgtaylor/huma/v2"
)

// obexEndpoints registers the endpoints for the "Obex" tagged endpoints.
func obexEndpoints(api huma.API, session bluetooth.Session) {
	startTransferEndpoint(api, session)
	cancelTransferEndpoint(api, session)
}

// cancelTransferEndpoint registers the path "/device/{address}/stop_file_transfer".
func cancelTransferEndpoint(api huma.API, session bluetooth.Session) {
	huma.Register(api, huma.Operation{
		OperationID: "file-transfer-stop",
		Method:      http.MethodGet,
		Path:        "/device/{address}/stop_file_transfer",
		Summary:     "Stop Transfers",
		Description: "Attempts to stop an ongoing file transfer session.",
		Tags:        []string{"File Transfer"},
	}, func(_ context.Context, input *struct {
		AddressInput
	},
	) (*struct{}, error) {
		return nil, session.Obex(input.Address).FileTransfer().CancelTransfer()
	})
}

// startTransferEndpoint registers the path "/device/{address}/start_file_transfer".
func startTransferEndpoint(api huma.API, session bluetooth.Session) {
	type SendFilesInput struct {
		Body struct {
			FilePaths []string `doc:"The full paths of files to be sent." example:"/music/music.mp3, /documents/doc.pdf" json:"file_paths" required:"true"`
		}
	}

	type QueuedFilesOutput struct {
		Body struct {
			Queued []bluetooth.FileTransferData `doc:"The files queued for transfer." json:"queued_files"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "file-transfer-start",
		Method:      http.MethodPost,
		Path:        "/device/{address}/start_file_transfer",
		Summary:     "Start Transfers",
		Description: "Attempts to send files to a device. If files are queued, monitor the `filetransfer` event in the `/events` stream for all ongoing file transfer events.",
		Tags:        []string{"File Transfer"},
	}, func(ctx context.Context, input *struct {
		AddressInput
		SendFilesInput
	},
	) (*QueuedFilesOutput, error) {
		if len(input.Body.FilePaths) == 0 {
			return nil, huma.Error422UnprocessableEntity("Empty filepath set provided.")
		}

		obexCall := session.Obex(input.Address)
		if err := obexCall.FileTransfer().CreateSession(ctx); err != nil {
			return nil, err
		}

		queued := &QueuedFilesOutput{}
		queued.Body.Queued = make([]bluetooth.FileTransferData, 0, len(input.Body.FilePaths))

		for _, file := range input.Body.FilePaths {
			select {
			case <-ctx.Done():
				return nil, obexCall.FileTransfer().CancelTransfer()
			default:
			}

			data, err := obexCall.FileTransfer().SendFile(file)
			if err != nil {
				return nil, err
			}

			queued.Body.Queued = append(queued.Body.Queued, data)
		}

		return queued, nil
	})
}
