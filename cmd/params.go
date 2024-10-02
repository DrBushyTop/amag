package cmd

import (
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"regexp"
)

const (
	KeyFile                     = "file"
	KeyMetric                   = "metric"
	KeyWorkspaceID              = "workspaceid"
	KeyScopeResourceID          = "scoperesourceid"
	KeyDataCollectionEndpoint   = "dataCollectionEndpoint"
	KeyDataCollectionStreamName = "dataCollectionStreamName"
	KeyDataCollectionRuleId     = "dataCollectionRuleId"
)

func bind(cmd *cobra.Command, keyName string, shortHand string, value string, usage string) error {
	cmd.Flags().StringP(keyName, shortHand, value, usage)
	_ = cmd.MarkFlagRequired(keyName)
	err := viper.BindPFlag(GetViperKey(cmd, keyName), cmd.Flags().Lookup(keyName))
	if err != nil {
		log.Error("Failed to bind flag", "name", keyName, "err", err)
		return err
	}
	return nil
}

func GetViperKey(cmd *cobra.Command, key string) string {
	return fmt.Sprintf("%s.%s", cmd.Name(), key)
}

func bindPersistent(cmd *cobra.Command, keyName string, shortHand string, value string, usage string) error {
	cmd.PersistentFlags().StringP(keyName, shortHand, value, usage)
	_ = cmd.MarkFlagRequired(keyName)
	err := viper.BindPFlag(keyName, cmd.Flags().Lookup(keyName))
	if err != nil {
		log.Error("Failed to bind flag", "name", keyName, "err", err)
		return err
	}
	return nil
}

func validateResourceId(resourceId string) error {
	unifiedPattern := `^/subscriptions/([a-f0-9\-]{36})/resourceGroups/([a-zA-Z0-9_\-\.]+)/providers/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+)(?:/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+))?(?:/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9_\-\.]+))?$`
	reUnified := regexp.MustCompile(unifiedPattern)

	if !reUnified.MatchString(resourceId) {
		return fmt.Errorf("invalid resourceId format. Metrics only support resource or subresource scope")
	}
	return nil
}