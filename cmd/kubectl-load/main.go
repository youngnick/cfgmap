package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/youngnick/kubectl-directory-output/internal/contexthelp"
	"github.com/youngnick/kubectl-directory-output/internal/signalhelp"
	"github.com/youngnick/kubectl-directory-output/pkg/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // required for GKE
)

func main() {

	stopCh := signalhelp.SetupHandler()
	ctx := contexthelp.WithStopCh(context.Background(), stopCh)
	cmd := cli.NewLoadCmd(ctx, os.Stdin, os.Stdout, os.Stderr)
	cobra.CheckErr(cmd.Execute())
}
