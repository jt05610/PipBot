/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"math"
	pb "pipbot/pipbot"
	"time"
)

func Do(bot *pb.PipBot, target *pb.Position) {
	dX := float64(target.X - bot.Current.X)
	dY := float64(target.Y - bot.Current.Y)
	dZ := float64(target.Z - bot.Current.Z)
	bot.GoTo(target)
	if (dX != 0) || (dY != 0) {
		dP := math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2))
		t := dP / bot.Rate
		time.Sleep(time.Duration(t) * time.Second)
	}
	if dZ != 0 {
		t := math.Abs(dZ / 10)
		time.Sleep(time.Duration(t*1000000) * time.Microsecond)
	}
}

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
		bot.Home()
		target := bot.Current
		target.Z = 85
		Do(bot, target)
		target = bot.Layout.Matrices[3].Cells[1][1]
		Do(bot, target)
		target.Z = 142
		Do(bot, target)
		target = bot.Layout.Matrices[2].Cells[2][2]
		Do(bot, target)
		target.Z = 142
		Do(bot, target)
		target = bot.Layout.Matrices[0].Cells[0][1]
		Do(bot, target)
		target.Z = 142
		Do(bot, target)
		target = bot.Layout.Matrices[1].Cells[5][9]
		Do(bot, target)
		target.Z = 142
		Do(bot, target)
		bot.Eject()
	},
}

func init() {
	rootCmd.AddCommand(tipCmd)

}
