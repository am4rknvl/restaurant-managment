package config

import (
	"log"
	"os"
)

// PaymentsConfig holds Telebirr/Fabric configuration loaded from environment variables.
type PaymentsConfig struct {
	MerchantAppID string
	FabricAppID   string
	ShortCode     string
	AppSecret     string
	PrivateKeyPEM string
}

var paymentsConfig PaymentsConfig

// Load reads and validates required environment variables. It should be called once at startup.
func Load() {
	paymentsConfig = PaymentsConfig{
		MerchantAppID: os.Getenv("TELEBIRR_MERCHANT_APP_ID"),
		FabricAppID:   os.Getenv("TELEBIRR_FABRIC_APP_ID"),
		ShortCode:     os.Getenv("TELEBIRR_SHORT_CODE"),
		AppSecret:     os.Getenv("TELEBIRR_APP_SECRET"),
		PrivateKeyPEM: os.Getenv("TELEBIRR_PRIVATE_KEY_PEM"),
	}

	// Do not fatally exit to keep non-payment features usable in dev; just warn if missing.
	if paymentsConfig.MerchantAppID == "" || paymentsConfig.FabricAppID == "" || paymentsConfig.ShortCode == "" || paymentsConfig.AppSecret == "" || paymentsConfig.PrivateKeyPEM == "" {
		log.Println("warning: missing Telebirr env vars; payment features may not work")
	}
}

// Payments returns a copy of the loaded PaymentsConfig.
func Payments() PaymentsConfig {
	return paymentsConfig
}


