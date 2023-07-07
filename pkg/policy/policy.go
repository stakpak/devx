package policy

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	log "github.com/sirupsen/logrus"
	"github.com/stakpak/devx/pkg/auth"
	"github.com/stakpak/devx/pkg/utils"
)

var policyNamePath = cue.ParsePath("$metadata.policy")

type GlobalPolicy struct {
	Name         string   `json:"name"`
	Environments []string `json:"environments"`
	PipelineJSON string
	IsEnforced   bool `json:"enforced"`
	IsDisabled   bool `json:"disabled"`
}
type GlobalPolicyData struct {
	Name         string   `json:"name"`
	Environments []string `json:"environments"`
	PipelineJSON string   `json:"pipeline"`
	IsEnforced   bool     `json:"enforced"`
	IsDisabled   bool     `json:"disabled"`
}

func Publish(configDir string, server auth.ServerConfig) error {
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return err
	}

	value, _, _ := utils.LoadProject(configDir, &overlays)
	codec := gocodec.New((*cue.Runtime)(value.Context()), nil)

	fieldIter, err := value.Fields()
	if err != nil {
		return err
	}
	for fieldIter.Next() {
		item := fieldIter.Value()
		policyNameValue := item.LookupPath(policyNamePath)
		if !policyNameValue.Exists() {
			continue
		}
		policyName, err := policyNameValue.String()
		if err != nil {
			return fmt.Errorf("invalid policy name %s", item.Path())
		}

		err = item.Validate(cue.Concrete(true))
		if err != nil {
			return fmt.Errorf("policy %s is not concrete", policyName)
		}

		policy := GlobalPolicy{}
		err = codec.Encode(item, &policy)
		if err != nil {
			return err
		}

		policyData := GlobalPolicyData(policy)
		response, err := utils.SendData(server, "policies", policyData)
		if err != nil {
			log.Debug(string(response))
			return err
		}

		log.Infof("Saved policy %s", policyName)
	}

	return nil
}
