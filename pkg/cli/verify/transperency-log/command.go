package transperencylog

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/devopstoday11/sigrun/pkg/config"
	"github.com/tidwall/pretty"

	rekorClient "github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "transperency-log [index]",
		Short: "Displays transperency log information for given log index",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) < 1 {
				return fmt.Errorf("log index is mandatory")
			}

			index, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			client, err := rekorClient.GetRekorClient(config.REKOR_URL)
			if err != nil {
				return err
			}

			req := entries.NewGetLogEntryByIndexParams()
			req.SetLogIndex(int64(index))
			resp, err := client.Entries.GetLogEntryByIndex(req)
			if err != nil {
				return err
			}

			encodedResp, err := json.Marshal(resp.GetPayload())
			if err != nil {
				return err
			}

			fmt.Println(string(pretty.Pretty(encodedResp)))

			return nil
		},
	}

	return cmd
}
