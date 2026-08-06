package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/ociregistry"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var ociClientOnce sync.Once
var ociClient ociregistry.OCIClienter

func ProvideOCIClient() ociregistry.OCIClienter {
	ociClientOnce.Do(func() {
		cfg := ProvideConfig()
		if !cfg.Clients.OCIClient.Enabled {
			logger.TechLog.Info(context.Background(), "OCI client is disabled, using test client")
			ociClient = ociregistry.NewTestClient()
		} else {
			var err error
			ociClient, err = ociregistry.NewClient(cfg)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide oci client: '%v'", err))
			}
		}
	})
	return ociClient
}
