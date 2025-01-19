package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	ac "github.com/bluetuith-org/api-native/api/appfeatures"
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/config"
	"github.com/bluetuith-org/api-native/api/eventbus"
	"github.com/bluetuith-org/api-native/platform"
	"github.com/bluetuith-org/bluerestd/endpoints"
	"github.com/danielgtaylor/huma/v2"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const (
	tcpURI   = "127.0.0.1:8888"
)

// These values are set at compile-time.
var (
	Version  = ""
	Revision = ""
)

// cmdError describes a command error.
type cmdError struct {
	spinner *pterm.SpinnerPrinter
	err     error
}

// Error returns the error as a string.
func (c cmdError) Error() string {
	return c.err.Error()
}

// New returns a new application.
func New() *cli.App {
	return cliApp()
}

// newCmdError returns a new cmdError.
func newCmdError(sp *pterm.SpinnerPrinter, err error) cmdError {
	return cmdError{sp, err}
}

// cliApp registers and returns the application instance.
func cliApp() *cli.App {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Fprintf(cCtx.App.Writer, "%s (%s)\n", Version, Revision)
	}

	return &cli.App{
		Name:                   "bluerestd",
		Usage:                  "Bluetooth REST API daemon.",
		Version:                Version + " (" + Revision + ")",
		Description:            "A Bluetooth daemon that provides a REST API to control Bluetooth Classic functionalities.\nNote that, certain endpoints may be disabled, depending on whether the underlying implementation supports certain functions. ",
		DefaultCommand:         "bluerestd launch",
		Copyright:              "(c) bluetuith-org.",
		Compiled:               time.Now(),
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Suggest:                true,
		Commands: []*cli.Command{
			{
				Name:  "openapi",
				Usage: "Generate an OpenAPI spec documentation.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "format",
						Usage:       "The format to be used to output the OpenAPI spec (json or yaml).",
						DefaultText: "json",
						Value:       "json",
						Aliases:     []string{"f"},
						EnvVars:     []string{"BRESTD_OAPI_FORMAT"},
					},
					&cli.StringFlag{
						Name:        "version",
						Usage:       "The version of the OpenAPI spec to be output (one of '3.1' or '3.0.3').",
						DefaultText: "3.0.3",
						Value:       "3.0.3",
						Aliases:     []string{"v"},
						EnvVars:     []string{"BRESTD_OAPI_VERSION"},
					},
				},
				Action: cmdOpenAPI,
			},
			{
				Name:        "launch",
				Usage:       "Start the daemon and listen for incoming API requests.",
				Description: "This subcommand requires either of the 'tcp-address' or 'unix-socket' options to be set.\nIf both options are empty, the default TCP address is used to listen for incoming API requests.",
				Flags: []cli.Flag{
					&cli.DurationFlag{
						Name:        "auth-timeout",
						Usage:       "The authentication timeout for device pairing and file transfer (in seconds).",
						Required:    false,
						DefaultText: "10",
						Value:       10,
						Aliases:     []string{"t"},
						EnvVars:     []string{"BRESTD_AUTHTIMEOUT"},
					},
					&cli.StringFlag{
						Name:        "tcp-address",
						Usage:       "The TCP address to listen on for API operations.",
						Required:    false,
						DefaultText: tcpURI,
						Value:       tcpURI,
						Aliases:     []string{"a"},
						EnvVars:     []string{"BRESTD_TCPADDR"},
					},
					&cli.StringFlag{
						Name:        "unix-socket",
						Usage:       "The UNIX socket path to listen on for API operations.\nIn this case, the 'http+unix' protocol is used, and clients can connect using this protocol.\nNote that the socket does not need to be created prior to using this option, it will be created automatically.\nIf the socket exists, it will return an error. For example, to connect to the socket via 'curl', use:\n curl --unix-socket /tmp/bluerestd.sock http://localhost/<endpoint>.",
						DefaultText: "/tmp/bluerestd.sock",
						Aliases:     []string{"s"},
						EnvVars:     []string{"BRESTD_SOCKET"},
					},
				},
				Action: cmdStart,
			},
		},
		ExitErrHandler: func(_ *cli.Context, err error) {
			if err == nil {
				return
			}

			cmdErr := &cmdError{}
			if !errors.As(err, cmdErr) {
				pterm.Error.Println(err)

				return
			}

			if cmdErr.err != nil {
				errorSpinner(cmdErr.spinner, cmdErr.err)
			}
		},
	}
}

// cmdStart handles the 'launch' command.
func cmdStart(cliCtx *cli.Context) error {
	if cliCtx.IsSet("tcp-address") && cliCtx.IsSet("unix-socket") {
		return fmt.Errorf("%s", "Only one of '--tcp-address' or '--unix-socket' must be specified.")
	}

	spinner := infoSpinner("Starting session")

	tcpaddr := cliCtx.String("tcp-address")
	sockpath := cliCtx.String("unix-socket")
	proto, addr := "tcp", tcpaddr

	if sockpath != "" {
		proto, addr = "unix", sockpath
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		return newCmdError(spinner, fmt.Errorf("Cannot listen on %s '%s': %w", proto, addr, err))
	}

	session, features, err := newSession(cliCtx)
	if err != nil {
		return newCmdError(spinner, err)
	}

	router := http.NewServeMux()
	endpoints.Register(router, session, features)

	err = serve(listener, router, spinner)
	if e := session.Stop(); e != nil {
		err = errors.Join(err, fmt.Errorf("Session shutdown error: %w", e))
	}

	if err == nil {
		spinner.Info("Exited.")
	}

	return newCmdError(spinner, err)
}

// cmdOpenAPI handles the 'openapi' command.
func cmdOpenAPI(cliCtx *cli.Context) error {
	oldFormat := false
	apifn := func() *huma.OpenAPI {
		api := endpoints.Register(http.NewServeMux(), nil, ac.MergedFeatureSet())

		return api.OpenAPI()
	}

	var (
		b   []byte
		err error
	)

	switch v := cliCtx.String("version"); v {
	case "3.0.3":
		oldFormat = true
	case "3.1":
		oldFormat = false
	default:
		err = fmt.Errorf("Invalid OpenAPI version: %s", v)

		goto Done
	}

	switch f := cliCtx.String("format"); f {
	case "json":
		if oldFormat {
			b, err = apifn().Downgrade()

			break
		}

		b, err = apifn().MarshalJSON()

	case "yaml":
		if oldFormat {
			b, err = apifn().DowngradeYAML()

			break
		}

		b, err = apifn().YAML()

	default:
		err = fmt.Errorf("Invalid OpenAPI format: %s", f)

		goto Done
	}

	fmt.Println(string(b))

Done:
	return err
}

// serve starts the HTTP server.
func serve(listener net.Listener, router *http.ServeMux, spinner *pterm.SpinnerPrinter) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create a new router & API.
	errchan := make(chan error, 1)
	server := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     router,
	}

	go func() {
		cstr, astr := "TCP address", listener.Addr().String()
		cstyle := pterm.NewStyle(pterm.Underscore, pterm.Bold, pterm.FgDefault).Sprint
		astyle := pterm.NewStyle(pterm.BgLightBlue, pterm.Bold).Sprint

		if listener.Addr().Network() == "unix" {
			cstr = "UNIX socket"
		} else {
			astyle = pterm.NewRGBStyle(
				pterm.NewRGB(0, 0, 0), pterm.NewRGB(0, 128, 255),
			).Sprint
		}

		printNote("%s", "Access the '/docs' endpoint with a web browser for an interactive API viewer.")
		printNote("%s", "The documentation for the OpenAPI specification is rendered offline.")
		newline()

		updateSpinner(spinner, "Listening on %s %s ...", cstyle(cstr), astyle(astr))

		// Start the server!
		if err := server.Serve(listener); err != nil {
			errchan <- fmt.Errorf(
				"Server startup error on %s '%s': %w",
				listener.Addr().Network(), listener.Addr().String(), err,
			)

			return
		}
	}()

	var err error

	select {
	case <-ctx.Done():
	case err = <-errchan:
	}

	clearSpinner(spinner)
	updateSpinner(spinner, "Exiting, please wait...")

	if e := server.Shutdown(ctx); e != nil && !errors.Is(err, context.Canceled) {
		err = errors.Join(err, fmt.Errorf("Server shutdown error: %w", e))
	}

	return err
}

// newSession initializes and returns a new session.
func newSession(cliCtx *cli.Context) (bluetooth.Session, ac.FeatureSet, error) {
	eventbus.DisableEvents()

	cfg := config.New()
	cfg.AuthTimeout = cliCtx.Duration("auth-timeout") * time.Second

	session, pinfo := platform.Session()

	features, err := session.Start(endpoints.NewAuthorizer(), cfg)
	if err != nil {
		return nil, features, fmt.Errorf("Session initialization error: %w", err)
	}

	if cerrs, ok := features.Errors.Exists(); ok {
		cstyle := pterm.NewStyle(pterm.FgLightYellow, pterm.Bold)
		estyle := pterm.NewStyle(pterm.FgRed, pterm.Bold, pterm.Underscore)

		printWarn("Feature(s) not available:")

		nodes := make([]pterm.TreeNode, 0, len(cerrs))
		for c, err := range cerrs {
			nodes = append(nodes, pterm.TreeNode{
				Text: cstyle.Sprintf(
					"'%s' -> %s",
					c.String(), estyle.Sprint(err.FeatureErrors.Error()),
				),
			})
		}

		pterm.DefaultTree.WithRoot(pterm.TreeNode{Children: nodes}).Render()
	}

	printInfo("Session initialized.")
	printInfo("Bluetooth stack: %s, OS: %s", pinfo.Stack.String(), pinfo.OS)
	newline()

	return session, features, nil
}
