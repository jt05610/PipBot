/*
Copyright © 2023 Jonathan Taylor <jonrtaylor12@gmail.com>
*/

package cmd

import (
	"context"
	"github.com/spf13/cobra"
	pb "pipbot/pipbot"
)

// homeCmd represents the home command
var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "homes the bot",
	Long:  `Sends G28. Be wary of clearances and things hitting other things!!`,
	Run: func(cmd *cobra.Command, args []string) {
		bot := pb.NewPipBot(pb.Port, pb.Baud)
		bot.Rate = 500
		ctx := context.Background()
		go bot.Sender(ctx)
		tips := pb.NewMatrix("Tips", &pb.Position{
			X: 165,
			Y: 103.5,
			Z: 73.5,
		}, 173.5-165, 173.5-165, 12, 8,
		)
		_ = bot.Listen(ctx)
		bot.Home()
		bot.Clear()
		bot.GoTo(tips.Home)
		p := tips.Home
		p.Z = 142
		bot.GoTo(p)
		bot.Eject()
	},
}

func init() {
	rootCmd.AddCommand(homeCmd)
}
