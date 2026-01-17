package config

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ChenBigdata421/jxt-core/sdk/config"
)

var (
	configYml string
	StartCmd  = &cobra.Command{
		Use:     "config",
		Short:   "Get Application config info",
		Example: "go-admin config -c config/settings.yml",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/settings.yml", "Start server with provided configuration file")
}

func run() {
	config.Setup(configYml)

	applicationConfig, errs := json.MarshalIndent(config.ApplicationConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("application:", string(applicationConfig))

	loggerConfig, errs := json.MarshalIndent(config.LoggerConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("logger:", string(loggerConfig))

	httpConfig, errs := json.MarshalIndent(config.HttpConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("http:", string(httpConfig))

	etcdConfig, errs := json.MarshalIndent(config.EtcdConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("etcd:", string(etcdConfig))

	grpcConfig, errs := json.MarshalIndent(config.GrpcConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("grpc:", string(grpcConfig))

	jwtConfig, errs := json.MarshalIndent(config.JwtConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("jwt:", string(jwtConfig))

	// todo 需要兼容
	databaseConfig, errs := json.MarshalIndent(config.DatabaseConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("database:", string(databaseConfig))

	queueConfig, errs := json.MarshalIndent(config.QueueConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("queue:", string(queueConfig))

	tenantConfig, errs := json.MarshalIndent(config.TenantsConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		fmt.Println(errs.Error())
	}
	fmt.Println("tenant:", string(tenantConfig))

}
