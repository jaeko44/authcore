package app

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"authcore.io/authcore/pkg/messageencryptor"

	"github.com/spf13/cobra"
)

var secret string
var operation string
var data string
var purpose string

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug commands",

	Args: cobra.MinimumNArgs(1),
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt command",

	Run: func(cmd *cobra.Command, args []string) {
		encryptOperation()
	},
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt command",

	Run: func(cmd *cobra.Command, args []string) {
		decryptOperation()
	},
}

func encryptOperation() {
	bSecret, err := hex.DecodeString(secret)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	keyGenerator := messageencryptor.NewKeyGenerator(bSecret)
	messageEncryptor, err := messageencryptor.NewMessageEncryptor(
		keyGenerator.Derive(
			"FieldEncryptor/Xsalsa20Poly1305",
			messageencryptor.CipherXsalsa20Poly1305.KeyLength(),
		),
		messageencryptor.CipherXsalsa20Poly1305,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if data == "" {
		reader := bufio.NewReader(os.Stdin)
		for {
			data, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			encrypt(messageEncryptor, data)
		}
	} else {
		encrypt(messageEncryptor, data)
	}
}

func decryptOperation() {
	bSecret, err := hex.DecodeString(secret)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	keyGenerator := messageencryptor.NewKeyGenerator(bSecret)
	messageEncryptor, err := messageencryptor.NewMessageEncryptor(
		keyGenerator.Derive(
			"FieldEncryptor/Xsalsa20Poly1305",
			messageencryptor.CipherXsalsa20Poly1305.KeyLength(),
		),
		messageencryptor.CipherXsalsa20Poly1305,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if data == "" {
		reader := bufio.NewReader(os.Stdin)
		for {
			data, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			decrypt(messageEncryptor, data)
		}
	} else {
		decrypt(messageEncryptor, data)
	}
}

func init() {
	encryptCmd.Flags().StringVarP(&secret, "secret", "s", "", "The secret. Should be hex encoded.")
	encryptCmd.Flags().StringVarP(&data, "data", "d", "", "The data to be encrypt/decrypted. Should be urlsafe-base64 encoded.")
	encryptCmd.Flags().StringVarP(&purpose, "purpose", "p", "purpose", "")
	encryptCmd.MarkFlagRequired("secret")
	encryptCmd.MarkFlagRequired("purpose")
	debugCmd.AddCommand(encryptCmd)

	decryptCmd.Flags().StringVarP(&secret, "secret", "s", "", "The secret. Should be hex encoded.")
	decryptCmd.Flags().StringVarP(&data, "data", "d", "", "The data to be encrypt/decrypted. Should be urlsafe-base64 encoded.")
	decryptCmd.Flags().StringVarP(&purpose, "purpose", "p", "purpose", "")
	decryptCmd.MarkFlagRequired("secret")
	decryptCmd.MarkFlagRequired("purpose")
	debugCmd.AddCommand(decryptCmd)
}

// $ go run authcore.io/authcore/cmd/authcorectl debug encrypt -s 855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268 -p secrets.secret -d 1337
// C6ow76xTkop8BBkoEjMV3V1rNNgp7DIRbjKMNEo9edy2J5PT3x9dg9EwkQNJTkzIxmN8Q9To3R6BE2zqagg

// $ go run authcore.io/authcore/cmd/authcorectl debug decrypt -s 855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268 -p secrets.secret -d C6ow76xTkop8BBkoEjMV3V1rNNgp7DIRbjKMNEo9edy2J5PT3x9dg9EwkQNJTkzIxmN8Q9To3R6BE2zqagg
// 1337

func encrypt(messageEncryptor *messageencryptor.MessageEncryptor, data string) {
	plaintext, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ciphertext, err := messageEncryptor.Encrypt(plaintext, []byte(purpose))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(ciphertext)
}

func decrypt(messageEncryptor *messageencryptor.MessageEncryptor, data string) {
	bPlaintext, err := messageEncryptor.Decrypt(data, []byte(purpose))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	plaintext := base64.RawURLEncoding.EncodeToString(bPlaintext)
	fmt.Println(plaintext)
}
