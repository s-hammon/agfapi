package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"

	"github.com/s-hammon/agfapi/pkg/agfa"
	"github.com/s-hammon/p"
	"github.com/spf13/cobra"
)

var (
	user     string
	pass     string
	baseUrl  string
	clientId string

	client *agfa.Client
)

var rootCmd = &cobra.Command{
	Use:   "agfapi",
	Short: "Get resources from the AGFA FHIR API",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if _, err = url.ParseRequestURI(baseUrl); err != nil {
			return fmt.Errorf("invalid base url: %v", err)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&user, "username", "u", "", "username for session-based login")
	rootCmd.PersistentFlags().StringVarP(&pass, "password", "p", "", "password for session-based login")
	rootCmd.PersistentFlags().StringVar(&clientId, "client-id", "", "client id for session-based login")

	user = p.Coalesce(user, os.Getenv("AGFA_USER"))
	pass = p.Coalesce(pass, os.Getenv("AGFA_PASS"))
	baseUrl = os.Getenv("AGFA_URL")
	clientId = p.Coalesce(clientId, os.Getenv("AGFA_CLIENT"))
}

func Execute(args []string, in io.Reader, out, err io.Writer) int {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		} else {
			return 1
		}
	}

	return 0
}
