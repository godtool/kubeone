package main

import (
	"embed"
	"runtime"

	_ "github.com/KubeOperator/kubepi/cmd/server/docs"
	"github.com/KubeOperator/kubepi/pkg/network/ip"
	_ "github.com/KubeOperator/kubepi/service/model/v1/cluster"
	_ "github.com/KubeOperator/kubepi/service/model/v1/clusterrepo"
	_ "github.com/KubeOperator/kubepi/service/model/v1/docs"
	_ "github.com/KubeOperator/kubepi/service/model/v1/imagerepo"
	_ "github.com/KubeOperator/kubepi/service/model/v1/role"
	_ "github.com/KubeOperator/kubepi/service/model/v1/user"
	"github.com/KubeOperator/kubepi/service/route"
	"github.com/KubeOperator/kubepi/service/server"
	"github.com/spf13/cobra"
	_ "k8s.io/api/rbac/v1"
)

//go:generate swag init

//swag init -g "cmd/server/main.go" -o "cmd/server/docs" --parseDependency --parseInternal --parseDepth 2

var (
	configPath     string
	serverBindHost string
	serverBindPort int
)

//go:embed web/kubepi
var embedWebKubePi embed.FS

//go:embed web/dashboard
var embedWebDashboard embed.FS

//go:embed web/terminal
var embedWebTerminal embed.FS

//go:embed script/darwin/init-kube.sh
var webkubectlEntrypointDarwin string

//go:embed script/linux/init-kube.sh
var webkubectlEntrypointLinux string

//go:embed helper/ip/qqwry.dat
var IpCommonDictionary []byte

func init() {
	RootCmd.Flags().StringVar(&serverBindHost, "server-bind-host", "", "kubepi bind address")
	RootCmd.Flags().IntVar(&serverBindPort, "server-bind-port", 0, "kubepi bind port")
	RootCmd.Flags().StringVarP(&configPath, "config-path", "c", "", "config file path")
}

var RootCmd = &cobra.Command{
	Use:   "kubepi-server",
	Short: "A dashboard for kubernetes",
	RunE: func(cmd *cobra.Command, args []string) error {
		server.EmbedWebDashboard = embedWebDashboard
		server.EmbedWebTerminal = embedWebTerminal
		server.EmbedWebKubePi = embedWebKubePi
		if runtime.GOOS == "darwin" {
			server.WebkubectlEntrypoint = webkubectlEntrypointDarwin
		} else {
			server.WebkubectlEntrypoint = webkubectlEntrypointLinux
		}
		ip.IpCommonDictionary = IpCommonDictionary
		return server.Listen(route.InitRoute,
			server.WithCustomConfigFilePath(configPath),
			server.WithServerBindHost(serverBindHost),
			server.WithServerBindPort(serverBindPort))
	},
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		panic(err)
	}
}
