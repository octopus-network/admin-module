package keeper

import (
	"github.com/octopus-network/admin-module/x/adminmodule/types"
)

var _ types.QueryServer = Keeper{}
