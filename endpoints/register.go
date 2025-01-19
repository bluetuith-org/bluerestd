package endpoints

import (
	"net/http"

	ac "github.com/bluetuith-org/api-native/api/appfeatures"
	bluetooth "github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// Register selectively registers endpoints based on the available features of the session.
func Register(router *http.ServeMux, session bluetooth.Session, features ac.FeatureSet) huma.API {
	api := registerAPI(router)

	adapterEndpoints(api, session)
	deviceEndpoints(api, session)

	if features.Has(ac.FeatureSendFile, ac.FeatureReceiveFile) {
		obexEndpoints(api, session)
	}

	if features.Has(ac.FeatureNetwork) {
		networkEndpoints(api, session)
	}

	if features.Has(ac.FeatureMediaPlayer) {
		mediaPlayerEndpoints(api, session)
	}

	sessionEndpoints(api, session)

	return api
}

// registerAPI registers the endpoints to the router.
func registerAPI(router *http.ServeMux) huma.API {
	config := huma.DefaultConfig("", "")
	config.DocsPath = ""

	router.HandleFunc("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(staticHTML))
	})

	api := humago.New(router, config)
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		ctx.SetHeader("Retry-After", "10")
		next(ctx)
	})

	api.OpenAPI().Info = staticAPIInfo

	for tag, desc := range staticTagDescriptions {
		api.OpenAPI().Tags = append(api.OpenAPI().Tags, &huma.Tag{
			Name:        tag,
			Description: desc,
		})
	}

	return api
}
