package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/boeing-ai-gateway/cmd"
	"github.com/boeing-ai-gateway/boeing/apiclient"
	"github.com/boeing-ai-gateway/boeing/logger"
	"github.com/boeing-ai-gateway/boeing/pkg/cli/internal"
	"github.com/boeing-ai-gateway/boeing/pkg/cli/internal/localconfig"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Boeing struct {
	Debug  bool `usage:"Enable debug logging"`
	Client *apiclient.Client
}

func (a *Boeing) PersistentPre(*cobra.Command, []string) error {
	if os.Getenv("NO_COLOR") != "" || !term.IsTerminal(int(os.Stdout.Fd())) {
		color.NoColor = true
	}

	if a.Debug {
		logger.SetDebug()
	}

	if a.Client.Token == "" {
		a.Client = a.Client.WithTokenFetcher(internal.Token)
	}

	return nil
}

func New() *cobra.Command {
	root := &Boeing{
		Client: newClient(),
	}
	return cmd.Command(root,
		&Server{},
		&Login{root: root},
		&Logout{root: root},
		&MCP{root: root},
		&Scan{root: root},
		&Setup{root: root},
		&Skills{root: root},
		&Version{},
		&Daemon{},
	)
}

func (a *Boeing) Run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

func newClient() *apiclient.Client {
	baseURL := os.Getenv("BOEING_BASE_URL")
	if baseURL != "" {
		if appURL, err := internal.AppURLForAPIBaseURL(baseURL); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid BOEING_BASE_URL: %v\n", err)
			baseURL = ""
		} else {
			baseURL = localconfig.APIBaseURL(appURL)
		}
	}
	if baseURL == "" {
		if cfg, err := localconfig.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load Boeing config: %v\n", err)
		} else if cfg.DefaultURL != "" {
			baseURL = localconfig.APIBaseURL(cfg.DefaultURL)
		}
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080/api"
	}

	return &apiclient.Client{
		BaseURL: baseURL,
		Token:   os.Getenv(internal.TokenEnvVar),
	}
}
