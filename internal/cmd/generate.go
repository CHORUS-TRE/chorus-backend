package cmd

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/jwks"

	"github.com/spf13/cobra"
)

var generateJWKSPublicKey bool

var generatePrivateKeyCmd = &cobra.Command{
	Use:   "generate-private-key",
	Short: "generate a fresh private key for daemon.private_key",
	Long:  `generates a fresh P-256 EC private key, PEM-encoded, ready to paste into daemon.private_key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGeneratePrivateKey()
	},
}

var generateJWKSCmd = &cobra.Command{
	Use:   "generate-jwks",
	Short: "generate a fresh JWKS for services.openid_connect_provider.jwks",
	Long:  `generates a fresh RSA keypair wrapped in a JWKS document, ready to paste into services.openid_connect_provider.jwks (add --public-key for the Keycloak one-liner too)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateJWKS()
	},
}

func init() {
	generateJWKSCmd.Flags().BoolVar(&generateJWKSPublicKey, "public-key", false, "also print the public key as a one-line PEM body (for registering with an external Keycloak instance)")

	rootCmd.AddCommand(generatePrivateKeyCmd)
	rootCmd.AddCommand(generateJWKSCmd)
}

func runGeneratePrivateKey() error {
	key, err := crypto.GeneratePrivateKeyPEM()
	if err != nil {
		return fmt.Errorf("unable to generate private key: %w", err)
	}

	fmt.Print(key)
	return nil
}

func runGenerateJWKS() error {
	jwksJSON, publicKeyPEMBody, err := jwks.Generate()
	if err != nil {
		return fmt.Errorf("unable to generate JWKS: %w", err)
	}

	fmt.Println(jwksJSON)

	if generateJWKSPublicKey {
		fmt.Println("\nPUBLIC KEY (Keycloak one-liner body):")
		fmt.Println(publicKeyPEMBody)
	}
	return nil
}
