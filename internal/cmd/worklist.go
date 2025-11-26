package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/s-hammon/agfapi/pkg/agfa"
	"github.com/spf13/cobra"
)

var worklistCmd = &cobra.Command{
	Use:  "worklist",
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		default:
			return fmt.Errorf("invalid command %q", args[0])
		case "help":
			return cmd.Help()
		case "get":
			return nil
		}
	},
}

var (
	outputPath string
	out        io.Writer
)

func init() {
	rootCmd.AddCommand(worklistCmd)

	worklistCmd.AddCommand(getCmd)
	worklistCmd.PersistentFlags().StringVarP(&outputPath, "output-filepath", "o", "", "filepath to save results to (JSON)")
}

var getCmd = &cobra.Command{
	Use:     "get [bundle-id]",
	Args:    cobra.ExactArgs(1),
	PreRunE: requestPreRun,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		svcReqs, err := handleGetWorklist(args[0])
		if err != nil {
			return err
		}

		log.Printf("found %d results\n", len(svcReqs))
		prettyPrintJson(out, svcReqs)
		if closer, ok := out.(io.WriteCloser); ok {
			closer.Close()
		}
		return nil
	},
}

func handleGetWorklist(listId string) ([]agfa.ServiceRequest, error) {
	list, err := client.FetchListById(listId)
	if err != nil {
		return nil, fmt.Errorf("couldn't get bundle: %v", err)
	}

	svcReqs := make([]agfa.ServiceRequest, 0)
	ch := make(chan agfa.ServiceRequest, len(list.Entry))

	var (
		taskId string
		reqId  string
		wg     sync.WaitGroup
	)

	t1 := time.Now()
	for _, e := range list.Entry {
		if e.Item.IsTask() {
			wg.Go(func() {
				taskId = e.Item.ExtractTaskId()
				task, err := client.FetchTaskById(taskId)
				if err != nil {
					log.Printf("error fetching task ID %q: %v\n", taskId, err)
					return
				}
				reqId = task.ServiceRequestId()
				if reqId == "" {
					log.Printf("no reqId for task %q\n", taskId)
					return
				}

				svcReq, err := client.FetchServiceRequestById(reqId)
				if err != nil {
					log.Printf("error fetching service request ID %q: %v\n", reqId, err)
				}

				ch <- svcReq
			})

		}
	}

	go func() {
		wg.Wait()
		close(ch)
		log.Printf("elapsed time: %.2fs\n", time.Since(t1).Seconds())
	}()

	for svcReq := range ch {
		svcReqs = append(svcReqs, svcReq)
	}

	return svcReqs, nil
}

func prettyPrintJson(w io.Writer, obj any) {
	b, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Fprintln(w, string(b))
}

func pave() error {
	dir := filepath.Dir(outputPath)
	return os.MkdirAll(dir, 0o750)
}
