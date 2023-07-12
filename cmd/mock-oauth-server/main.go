package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/dosquad/mock-oauth-test-server/internal/mainconfig"
	"github.com/dosquad/mock-oauth-test-server/mockghauth"
	"github.com/na4ma4/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

//nolint:gochecknoglobals // cobra uses globals in main
var rootCmd = &cobra.Command{
	Use:  "mock-oauth-server",
	RunE: mainCommand,
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")

	_ = viper.BindEnv("load.code-file")

	_ = viper.BindEnv("load.code-file", "LOAD_CODE_FILE")
	_ = viper.BindEnv("load.clients-file", "LOAD_CLIENTS_FILE")
	_ = viper.BindEnv("load.tokens-file", "LOAD_TOKENS_FILE")
}

func main() {
	_ = rootCmd.Execute()
}

func mainCommand(_ *cobra.Command, _ []string) error {
	cfg := config.NewViperConfigFromViper(viper.GetViper(), "mock-server")

	logger, _ := cfg.ZapConfig().Build()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	baseURL, baseURLErr := url.Parse(fmt.Sprintf("http://%s", cfg.GetString("server.bind")))
	if baseURLErr != nil {
		return baseURLErr
	}

	svr := mockghauth.NewServer(baseURL, cfg)

	svr.AddClient("github-client-id", "github-client-secret")

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		ticker := time.NewTicker(time.Minute)
		for {
			ts := <-ticker.C
			svr.Reaper(ts)
		}
	})

	eg.Go(func() error {
		return svr.Run(ctx)
	})

	if err := eg.Wait(); err != nil {
		log.Printf("error: %s", err)
	}

	return nil
}
