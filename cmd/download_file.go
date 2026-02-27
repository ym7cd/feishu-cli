package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var downloadFileCmd = &cobra.Command{
	Use:   "download <file_token>",
	Short: "下载云空间文件",
	Long: `从云空间下载文件到本地。

参数:
  file_token    文件的 Token

选项:
  -o, --output  输出文件路径（默认使用当前目录下的文件名）

示例:
  # 下载文件到当前目录
  feishu-cli file download boxcnXXXXXXXXX

  # 下载文件到指定路径
  feishu-cli file download boxcnXXXXXXXXX -o /tmp/myfile.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		fileToken := args[0]
		outputPath, _ := cmd.Flags().GetString("output")

		if outputPath == "" {
			outputPath = fileToken
		}

		if err := client.DownloadFile(fileToken, outputPath); err != nil {
			return err
		}

		fmt.Printf("文件下载成功！\n")
		fmt.Printf("  文件 Token: %s\n", fileToken)
		fmt.Printf("  保存路径:   %s\n", outputPath)

		return nil
	},
}

func init() {
	fileCmd.AddCommand(downloadFileCmd)
	downloadFileCmd.Flags().StringP("output", "o", "", "输出文件路径")
}
