package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/s-hammon/agfapi/pkg/agfa"
	"github.com/s-hammon/p"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	queryParams []string

	caser = cases.Title(language.AmericanEnglish)
)

var requestCmd = &cobra.Command{
	Use:     "request [endpoint]",
	Args:    cobra.ExactArgs(1),
	PreRunE: requestPreRun,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		parts := strings.Split(args[0], "/")
		parts[0] = caser.String(parts[0])
		endpoint := strings.Join(parts, "/")
		params := p.StringDeserialize(queryParams)

		var res map[string]any
		if err = client.Get(endpoint, params, &res); err != nil {
			return fmt.Errorf("client.Get: %v", err)
		}

		prettyPrintJson(out, res)
		if closer, ok := out.(io.WriteCloser); ok {
			closer.Close()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(requestCmd)
	requestCmd.Flags().StringSliceVarP(&queryParams, "query-param", "q", []string{}, "specify query param (key=value)")
}

func requestPreRun(cmd *cobra.Command, args []string) (err error) {
	out = cmd.OutOrStdout()
	if err = checkOutPath(); err != nil {
		return err
	}

	return newClient()
}

func checkOutPath() (err error) {
	if outputPath != "" {
		if err = pave(); err != nil {
			return fmt.Errorf("couldn't create output file path directory: %v", err)
		}
		out, err = os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("couldn't create output file: %v", err)
		}
	}

	return nil
}

func newClient() (err error) {
	log.Println("logging in...")
	client, err = agfa.NewClient(baseUrl).Session(agfa.SessionParams{
		Username: user,
		Password: pass,
		ClientId: clientId,
	})
	if err != nil {
		return fmt.Errorf("agfa.NewClient: %v", err)
	}

	log.Println("client logged in")
	return nil
}
