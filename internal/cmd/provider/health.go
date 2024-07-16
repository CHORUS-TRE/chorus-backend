package provider

import (
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
)

var healthControllerOnce sync.Once
var healthController chorus.HealthServiceServer

func ProvideHealthController() chorus.HealthServiceServer {
	healthControllerOnce.Do(func() {
		healthController = v1.NewHealthController()
	})
	return healthController
}
