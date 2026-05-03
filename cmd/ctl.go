package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var ctlCmd = &cobra.Command{
	Use:                "ctl",
	Short:              "Run kubectl with KUBECONFIG=./kube_config.yaml",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubectl, err := exec.LookPath("kubectl")
		if err != nil {
			return err
		}
		c := exec.Command(kubectl, args...)
		c.Env = append(os.Environ(), "KUBECONFIG=./kube_config.yaml")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
