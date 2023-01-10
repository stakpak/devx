package policy

import "fmt"

func Publish(configDir string, telemetry string) error {
	if telemetry == "" {
		return fmt.Errorf("telemtry endpoint is required to publish policies")
	}

	return nil
}
