/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type WallResponse struct {
	Data []struct {
		Path string `json:"path"`
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "waldl_go",
	Short: "Wallpaper Downloader",
	Long:  "Downloads wallpaper from wallhaven.cc/",
	Args:  cobra.MaximumNArgs(1),
	Run:   rootRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.waldl_go.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func rootRun(cmd *cobra.Command, args []string) {
	wallLinks := getWallLinks(args)
	fmt.Println(wallLinks)
}

func getWallLinks(args []string) []string {
	wallUrl := "https://wallhaven.cc/api/v1/search"
	req, err := http.NewRequest(http.MethodGet, wallUrl, nil)
	if err != nil {
		log.Fatalln("Error occured when creating request", err)
	}

	// If there is an argument to query
	if len(args) > 0 {
		fmt.Println("Received query", args[0])
		q := req.URL.Query()
		q.Add("q", args[0])
		req.URL.RawQuery = q.Encode()
	}

	fmt.Println("Fetching first page of urls", req.URL.String())

	res, err := http.Get(req.URL.String())
	if err != nil {
		log.Fatalln("Error occured when requesting", req.URL.String(), err)
	}
	if res.StatusCode != 200 {
		log.Fatalln("Error occured, recieved response code", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("Error occured when reading response body", err)
	}
	var wallRes WallResponse
	if err := json.Unmarshal(body, &wallRes); err != nil {
		log.Fatalln("Error occured when marshalling response body to json", err)
	}

	wallLinks := make([]string, 0, len(wallRes.Data))
	for _, v := range wallRes.Data {
		wallLinks = append(wallLinks, v.Path)
	}
	return wallLinks
}
