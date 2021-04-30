package main

import (
	"github.com/youngnick/cfgmap/cmd/kubectl-cfgmap/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cli.Execute()
}
