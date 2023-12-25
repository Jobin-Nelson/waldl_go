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
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	downloadDir string
	numWall     int
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
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error occured, could not get home directory", err)
	}
	wallDir := filepath.Join(home, "Pictures", "wallpapers")
	if _, err := os.Stat(wallDir); os.IsNotExist(err) {
		log.Fatalln(wallDir, "does not exist")
	}
	rootCmd.PersistentFlags().StringVarP(
		&downloadDir,
		"dir",
		"d",
		filepath.Join(wallDir, time.Now().Format(time.DateOnly)),
		"directory to download wallpapers",
	)
	rootCmd.PersistentFlags().IntVarP(
		&numWall,
		"number",
		"n",
		0,
		"Number of wallpapers to download",
	)
}

func rootRun(cmd *cobra.Command, args []string) {
	fmt.Println("Downloading to", downloadDir)
	c := http.DefaultClient
	wallLinks := getWallLinks(c, args)
	if numWall == 0 {
		numWall = len(wallLinks)
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < numWall; i++ {
		wg.Add(1)
		go downloadWall(c, wallLinks[i], wg)
	}
	wg.Wait()

	fmt.Println("Wallpapers downloaded to", downloadDir)
}

func getWallLinks(c *http.Client, args []string) []string {
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

	res, err := c.Do(req)
	if err != nil {
		log.Fatalln("Error occured when requesting", req.URL.String(), err)
	}
	defer res.Body.Close()
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

func downloadWall(c *http.Client, l string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Fetching wallpaper
	res, err := c.Get(l)
	if err != nil {
		fmt.Println("Not able to download", l)
		return
	}
	defer res.Body.Close()

	// Creating file
	downloadFile := filepath.Join(downloadDir, filepath.Base(l))
	f, err := os.OpenFile(downloadFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Not able to create file", downloadFile)
		return
	}
	defer f.Close()

	// Copying wallpaper to file
	if _, err := io.Copy(f, res.Body); err != nil {
		fmt.Println("Not able to write to file", downloadFile)
		return
	}

	fmt.Println("Downloaded", downloadFile)
}
