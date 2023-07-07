package xray

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/stakpak/devx/pkg/auth"
	"github.com/stakpak/devx/pkg/utils"
)

type XRayRequest struct {
	Project string            `json:"project"`
	Source  map[string]string `json:"source"`
}
type XRayResponse struct {
	Project string            `json:"project"`
	Source  map[string]string `json:"source"`

	State        string    `json:"state"`
	Result       string    `json:"result"`
	Dependencies []XRayDep `json:"dependencies"`
	Environments []XRayEnv `json:"environments"`
	Languages    []string  `json:"languages"`
}
type XRayEnv struct {
	Name string `json:"name"`
}
type XRayDep struct {
	Name string `json:"name"`
}

func Run(configDir string, server auth.ServerConfig) error {
	log.Info("👁️  Scanning your source code...")
	overlay, err := scan(configDir)
	if err != nil {
		return err
	}

	item := XRayRequest{
		Project: "none",
		Source:  overlay,
	}

	log.Info("🧠 Analysing data...this will take a couple of mins...")
	data, err := utils.SendData(server, "xray", item)
	if err != nil {
		log.Debug(string(data))
		return err
	}
	resp := make(map[string]string)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	id := resp["id"]
	data, err = utils.GetData(server, "xray", &id, nil)
	if err != nil {
		log.Debug(string(data))
		return err
	}

	xrayResp := XRayResponse{}
	err = json.Unmarshal(data, &xrayResp)
	if err != nil {
		return err
	}

	log.Info("🤓 Analysis results")
	deps := []string{}
	for _, dep := range xrayResp.Dependencies {
		deps = append(deps, dep.Name)
	}

	tableData := [][]string{
		{"Languages", strings.Join(xrayResp.Languages, " ")},
		{"Dependencies", strings.Join(deps, " ")},
		{"Environment variables", fmt.Sprint(len(xrayResp.Environments))},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(tableData)
	table.Render()

	if xrayResp.State != "Success" {
		return fmt.Errorf("unexpected devx-ray response status %s", xrayResp.State)
	}

	log.Info("🏭 Generating your stack...")
	err = os.WriteFile("stack.gen.cue", []byte(xrayResp.Result), 0700)
	if err != nil {
		return err
	}

	log.Info("Created your stack at \"stack.gen.cue\"")
	return nil
}
