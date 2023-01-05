package cmd

import (
	"io"
	"os"
	"strings"

	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-rod/rod"
)

var cfgFile string

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func readStdin() string {
	bytes, err := io.ReadAll(os.Stdin)
	if err == nil {
		return string(bytes)
	}

	fmt.Printf("Could not read from stdin\n")
	os.Exit(1)
	return ""
}

func isStdin() bool {
	stat, _ := os.Stdin.Stat()

	return (stat.Mode() & os.ModeCharDevice) == 0
}

var expiresOptions = []string{"5min", "10min", "1hour", "1day", "1week", "1month", "1year", "never"}

var rootCmd = &cobra.Command{
	Use: `privatebin "string for privatebin"... [flags]`,
	Example: `privatebin "encrypt this string" --expires 1day --burn --password Secret
	cat textfile | privatebin --expires 5min --url https://privatebin.net
echo "hello\nworld" | privatebin --expires never -B`,
	Short:   "CLI access to privatebin",
	Version: "1.0.1",
	Run: func(cmd *cobra.Command, args []string) {

		url := viper.GetViper().GetString("url")

		if len(url) == 0 {
			fmt.Println("--url is required")
			os.Exit(1)
		}

		expires := viper.GetViper().GetString("expires")

		if !contains(expiresOptions, expires) {
			fmt.Println(fmt.Sprintf("--expires can only be one of %s\n", strings.Join(expiresOptions[:], ", ")))
			os.Exit(1)
		}

		stringToEncrypt := ""

		if isStdin() {
			stringToEncrypt = readStdin()
		} else {
			if len(args) == 0 {
				fmt.Println("No arguments supplied")
				os.Exit(1)
			}
			stringToEncrypt = strings.Join(args[:], " ")
		}

		page := rod.New().MustConnect().MustPage(url).MustWindowFullscreen()

		defer page.MustClose()

		burn := viper.GetViper().GetBool("burn")
		deleteLink := viper.GetViper().GetBool("delete")

		if cmd.Flags().Changed("password") {
			password, err := cmd.Flags().GetString("password")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			page.MustElement("#passwordinput").MustInput(password)
		}

		el := page.MustElement(`input[name="burnafterreading"]`)
		if el.MustProperty("checked").Bool() != burn {
			el.MustClick()
		}

		// Open the menu to select expiration time
		page.MustElement("#expiration").MustClick()

		page.MustElement(`[data-expiration="` + expires + `"]`).MustClick()

		page.MustElement("#message").MustInput(stringToEncrypt)

		page.MustElement("#sendbutton").MustClick()

		secretUrl := page.MustElement("#pasteurl").MustText()

		if !deleteLink {
			fmt.Println(secretUrl)
		} else {
			deleteUrl := page.MustElementR("a", "Delete data").MustAttribute("href")

			fmt.Println("Secret URL: " + secretUrl)
			fmt.Println(fmt.Sprintf("Delete URL: %s", *deleteUrl))
		}

	},
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.privatebin.yaml or $PWD/.privatebin.yaml)")
	rootCmd.PersistentFlags().String("expires", "5min", fmt.Sprintf("How long the snippet should live\n%s", strings.Join(expiresOptions[:], ", ")))
	rootCmd.PersistentFlags().BoolP("burn", "B", false, "Burn after reading")
	rootCmd.PersistentFlags().BoolP("delete", "D", false, "Show delete link")
	rootCmd.PersistentFlags().String("password", "", "Password for the snippet")
	rootCmd.PersistentFlags().String("url", "", "URL to privatebin app")

	viper.BindPFlag("expires", rootCmd.PersistentFlags().Lookup("expires"))
	viper.BindPFlag("burn", rootCmd.PersistentFlags().Lookup("burn"))
	viper.BindPFlag("delete", rootCmd.PersistentFlags().Lookup("delete"))
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".privatebin" (without extension).
		viper.AddConfigPath(home)
		// Search current dir
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".privatebin")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
