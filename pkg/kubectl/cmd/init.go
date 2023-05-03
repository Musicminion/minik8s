package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Execute() {
	if err := commnads.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}

var commnads = &cobra.Command{
	Use:   "kubectl",
	Short: "Kubectl is a tool for controlling minik8s cluster.",
	Long:  `Kubectl is a tool for controlling minik8s cluster. To see the help of a specific command, use: kubectl [command] --help`,
	Run:   runRoot,
}

func init() {
	commnads.AddCommand(applyCmd)
}

func runRoot(cmd *cobra.Command, args []string) {

	// Reach here if no args
	fmt.Printf("execute %s args:%v \n", cmd.Name(), args)
	fmt.Println("kubectl is for better control of minik8s")
	fmt.Println(cmd.UsageString())
}
