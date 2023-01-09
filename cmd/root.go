package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"privatebin/utils"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
)

var cfgFile string

var expiresOptions = []string{"5min", "10min", "1hour", "1day", "1week", "1month", "1year", "never"}
var outputOptions = []string{"simple", "rich", "json"}
var formatOptions = []string{"plain", "code", "md"}

func flagErrorMessage(flag string) error {
	return fmt.Errorf("%s is required", flag)
}

func flagErrorOption(flag string, options []string) error {
	return fmt.Errorf("%s can only be on of %s", flag, strings.Join(options[:], ", "))
}

func flagValidation() (string, string, string, string) {
	if len(viper.GetViper().GetString("url")) == 0 {
		fmt.Println(flagErrorMessage("--url"))
		os.Exit(1)
	}

	if !utils.Contains(outputOptions, viper.GetViper().GetString("output")) {
		fmt.Println(flagErrorOption("--output", outputOptions))
		os.Exit(1)
	}

	if !utils.Contains(expiresOptions, viper.GetViper().GetString("expires")) {
		fmt.Println(flagErrorOption("--expires", expiresOptions))
		os.Exit(1)
	}

	if !utils.Contains(formatOptions, viper.GetViper().GetString("format")) {
		fmt.Println(flagErrorOption("--format", formatOptions))
		os.Exit(1)
	}

	var formats = map[string]string{"plain": "plaintext", "code": "syntaxhighlighting", "md": "markdown"}

	return viper.GetViper().GetString("url"),
		viper.GetViper().GetString("output"),
		viper.GetViper().GetString("expires"),
		formats[viper.GetViper().GetString("format")]
}

var rootCmd = &cobra.Command{
	Use: `privatebin "string for privatebin"... [flags]`,
	Example: `privatebin "encrypt this string" --expires 1day --burn --password Secret
cat textfile | privatebin --url https://yourprivatebin.com`,
	Short:   "CLI access to privatebin",
	Version: utils.Version(),
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose, _ := cmd.Flags().GetBool("verbose")

		url, output, expires, format := flagValidation()

		burn := viper.GetViper().GetBool("burn")

		stringToEncrypt := ""
		if utils.IsStdin() {
			if isVerbose {
				fmt.Fprintln(os.Stderr, "Getting data from Stdin")
			}
			stringToEncrypt = utils.ReadStdin()
		} else {
			if len(args) == 0 || len(args[0]) == 0 {
				fmt.Println("No arguments supplied")
				os.Exit(1)
			}
			if isVerbose {
				fmt.Fprintln(os.Stderr, "Getting data from argument")
			}
			stringToEncrypt = args[0]
		}

		pasteContent, err := json.Marshal(&PasteContent{Paste: utils.StripANSI(stringToEncrypt)})
		if err != nil {
			panic(err)
		}

		if isVerbose {
			fmt.Fprintln(os.Stderr, "Generating master key")
		}
		masterKey, err := utils.GenRandomBytes(32)
		if err != nil {
			panic(err)
		}

		password, _ := cmd.Flags().GetString("password")

		if isVerbose {
			fmt.Fprintln(os.Stderr, "Encrypting data")
		}
		pasteData, err := encrypt(masterKey, password, pasteContent, format, burn)
		if err != nil {
			panic(err)
		}

		// Create a new Paste Request.
		pasteRequest := &PasteRequest{
			V:     2,
			AData: pasteData.adata(),
			Meta: PasteRequestMeta{
				Expire: expires,
			},
			CT: utils.Base64(pasteData.Data),
		}

		// Get the Request Body.
		body, err := json.Marshal(pasteRequest)
		if err != nil {
			panic(err)
		}

		// Create a new HTTP Client and HTTP Request.
		client := &http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			panic(err)
		}

		// Set the request headers.
		req.Header.Set("User-Agent", "privatebin-cli/"+cmd.Version+" (source; https://github.com/M0ter/privatebin-cli")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
		req.Header.Set("X-Requested-With", "JSONHttpRequest")

		// Run the http request.
		if isVerbose {
			fmt.Fprintln(os.Stderr, "Connecting to server:", url)
		}
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		// Close the request body once we are done.
		defer func() {
			if err := res.Body.Close(); err != nil {
				panic(err)
			}
			if isVerbose {
				fmt.Fprintln(os.Stderr, "Closing connection to server")
			}
		}()

		// Read the response body.
		response, err := ioutil.ReadAll(res.Body)
		if isVerbose {
			fmt.Fprintln(os.Stderr, "Reading response")
		}
		if err != nil {
			panic(err)
		}

		// Decode the response.
		pasteResponse := &PasteResponse{}
		if err := json.Unmarshal(response, &pasteResponse); err != nil {
			panic(err)
		}

		if isVerbose {
			fmt.Fprintln(os.Stderr, "Printing data")
		}
		secretUrl := fmt.Sprintf("%s%s#%s\n", url, pasteResponse.URL, base58.Encode(masterKey))

		deleteUrl := fmt.Sprintf("%s%s&deletetoken=%s\n", url, pasteResponse.URL, pasteResponse.DeleteToken)

		switch output {
		case "simple":
			fmt.Print(secretUrl)
			break
		case "rich":
			fmt.Print("Secret URL: " + secretUrl)
			fmt.Print(fmt.Sprintf("Delete URL: %s", deleteUrl))
			break
		case "json":
			fmt.Print(fmt.Sprintf(`{"url": "%s","delete_url": "%s"`, secretUrl, deleteUrl))
			break
		}

	},
}

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
	rootCmd.PersistentFlags().String("url", "https://privatebin.net", "URL to privatebin app")
	rootCmd.PersistentFlags().String("password", "", "Password for the paste")
	rootCmd.PersistentFlags().String("output", "simple", fmt.Sprintf("Output format of the returned data\n%s", strings.Join(outputOptions[:], ", ")))
	rootCmd.PersistentFlags().String("format", "plaintext", fmt.Sprintf("Paste format\n%s", strings.Join(formatOptions[:], ", ")))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	viper.BindPFlag("expires", rootCmd.PersistentFlags().Lookup("expires"))
	viper.BindPFlag("burn", rootCmd.PersistentFlags().Lookup("burn"))
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
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

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		isVerbose, _ := rootCmd.Flags().GetBool("verbose")
		if isVerbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

// SpecArray .
func (spec *PasteSpec) SpecArray() []interface{} {
	return []interface{}{
		spec.IV,
		spec.Salt,
		spec.Iterations,
		spec.KeySize,
		spec.TagSize,
		spec.Algorithm,
		spec.Mode,
		spec.Compression,
	}
}

func encrypt(master []byte, password string, message []byte, format string, burn bool) (*PasteData, error) {
	// Generate a initialization vector.
	iv, err := utils.GenRandomBytes(12)
	if err != nil {
		return nil, err
	}

	// Generate salt.
	salt, err := utils.GenRandomBytes(8)
	if err != nil {
		return nil, err
	}

	// Create the Paste Data and generate a key.
	paste := &PasteData{
		Formatter:        format,
		BurnAfterReading: burn,
		PasteSpec: &PasteSpec{
			IV:          utils.Base64(iv),
			Salt:        utils.Base64(salt),
			Iterations:  310000,
			KeySize:     256,
			TagSize:     128,
			Algorithm:   "aes",
			Mode:        "gcm",
			Compression: "none",
		},
	}

	master = append(master, []byte(password)...)

	key := pbkdf2.Key(master, salt, paste.Iterations, 32, sha256.New)

	adata, err := json.Marshal(paste.adata())
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	data := gcm.Seal(nil, iv, message, adata)

	paste.Data = data

	return paste, nil
}

// PasteRequest .
type PasteRequest struct {
	V     int              `json:"v"`
	AData []interface{}    `json:"adata"`
	Meta  PasteRequestMeta `json:"meta"`
	CT    string           `json:"ct"`
}

// PasteRequestMeta .
type PasteRequestMeta struct {
	Expire string `json:"expire"`
}

// PasteResponse .
type PasteResponse struct {
	Status      int    `json:"status"`
	ID          string `json:"id"`
	URL         string `json:"url"`
	DeleteToken string `json:"deletetoken"`
}

// PasteContent .
type PasteContent struct {
	Paste string `json:"paste"`
}

// PasteSpec .
type PasteSpec struct {
	IV          string
	Salt        string
	Iterations  int
	KeySize     int
	TagSize     int
	Algorithm   string
	Mode        string
	Compression string
}

// PasteData .
type PasteData struct {
	*PasteSpec
	Data             []byte
	Formatter        string
	BurnAfterReading bool
}

func (paste *PasteData) adata() []interface{} {
	var bool2int = map[bool]int8{false: 0, true: 1}
	return []interface{}{
		paste.SpecArray(),
		paste.Formatter,
		0,
		bool2int[paste.BurnAfterReading],
	}
}
