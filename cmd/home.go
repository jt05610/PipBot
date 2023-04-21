/*
Copyright © 2023 Jonathan Taylor <jonrtaylor12@gmail.com>
*/

package cmd

import (
	"github.com/spf13/cobra"
	pb "pipbot/pipbot"
)

// homeCmd represents the home command
var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "homes the bot",
	Long:  `Sends G28. Be wary of clearances and things hitting other things!!`,
	Run: func(cmd *cobra.Command, args []string) {
		bot := pb.NewPipBot(pb.Port, pb.Baud, 0)
		bot.Rate = 500
		bot.Home()
	},
}

func init() {
	rootCmd.AddCommand(homeCmd)
}
