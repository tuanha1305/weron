package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pojntfx/webrtcfd/pkg/wrtcmgr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	raddrFlag = "raddr"
)

var (
	errMissingAPIPassword = errors.New("missing API password")
	errMissingAPIUsername = errors.New("missing API username")
)

var managerListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"lis", "l", "ls"},
	Short:   "List persistent and ephermal communities",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if u := os.Getenv("API_USERNAME"); u != "" {
			if viper.GetBool(verboseFlag) {
				log.Println("Using username from API_USERNAME env variable")
			}

			viper.Set(apiUsernameFlag, u)
		}

		if u := os.Getenv("API_PASSWORD"); u != "" {
			if viper.GetBool(verboseFlag) {
				log.Println("Using password from API_PASSWORD env variable")
			}

			viper.Set(apiPasswordFlag, u)
		}

		return viper.BindPFlags(cmd.PersistentFlags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		if strings.TrimSpace(viper.GetString(apiPasswordFlag)) == "" {
			return errMissingAPIPassword
		}

		if strings.TrimSpace(viper.GetString(apiUsernameFlag)) == "" {
			return errMissingAPIUsername
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		manager := wrtcmgr.NewManager(
			viper.GetString(raddrFlag),
			viper.GetString(apiUsernameFlag),
			viper.GetString(apiPasswordFlag),
			ctx,
		)

		c, err := manager.ListCommunities()
		if err != nil {
			return err
		}

		w := csv.NewWriter(os.Stdout)
		defer w.Flush()

		if err := w.Write([]string{"id", "clients", "persistent"}); err != nil {
			return err
		}

		for _, community := range c {
			if err := w.Write([]string{community.ID, fmt.Sprintf("%v", community.Clients), fmt.Sprintf("%v", community.Persistent)}); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	managerListCmd.PersistentFlags().String(apiUsernameFlag, "admin", "Username for the management API (can also be set using the API_USERNAME env variable). Ignored if any of the OIDC parameters are set.")
	managerListCmd.PersistentFlags().String(apiPasswordFlag, "", "Password for the management API (can also be set using the API_PASSWORD env variable). Ignored if any of the OIDC parameters are set.")
	managerListCmd.PersistentFlags().String(raddrFlag, "https://webrtcfd-production.up.railway.app/", "Remote address")

	viper.AutomaticEnv()

	managerCmd.AddCommand(managerListCmd)
}