package system

import (
	coremodel "github.com/mejango/dia-go/models/core"
)

type NodeState struct {
	Network string        `json:"network"`
	Scene   coremodel.Ref `json:"scene"`
}

type KeyState struct {
	Revoked   string        `json:"revoked"`
	Name      string        `json:"name"`
	Network   string        `json:"network"`
	Secret    string        `json:"secret"`
	Scene     coremodel.Ref `json:"scene"`
	Principal coremodel.Ref `json:"principal"`
	Subject   coremodel.Ref `json:"subject"`
}
