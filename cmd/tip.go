/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/spf13/cobra"
	pb "pipbot/pipbot"
)

// tipCmd represents the tip command
var tipCmd = &cobra.Command{
	Use:   "tip",
	Short: "use to get tip",
	Long:  `tip gets tips `,
	Run: func(cmd *cobra.Command, args []string) {
		bot := pb.NewPipBot(pb.Port, pb.Baud, 0)
		bot.Rate = 500
		ctx := context.Background()
		_ = bot.Listen(ctx)
		bot.Init()
		bp := bot.Layout.Matrices[2]
		wp := bot.Layout.Matrices[1]

		bot.Transfer(bp.Cells[0][0], wp.Cells[1][0], 200, true)
	},
}

func init() {
	rootCmd.AddCommand(tipCmd)

}
