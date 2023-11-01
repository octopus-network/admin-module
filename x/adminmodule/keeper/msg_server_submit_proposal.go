package keeper

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"admin-module/x/adminmodule/types"
)

func (k msgServer) SubmitProposal(goCtx context.Context, msg *types.MsgSubmitProposal) (*types.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AdminKey))
	storeCreator := store.Get([]byte(msg.Proposer))
	if storeCreator == nil {
		return nil, fmt.Errorf("proposer %s must be admin to submit proposals to admin-module", msg.Proposer)
	}

	content := msg.GetContent()
	if !k.Keeper.IsProposalTypeWhitelisted(content) {
		return nil, errors.New("proposal content is not whitelisted")
	}

	proposal, err := k.Keeper.SubmitProposal(ctx, content)
	if err != nil {
		return nil, err
	}

	defer telemetry.IncrCounter(1, types.ModuleName, "proposal")

	submitEvent := sdk.NewEvent(types.EventTypeSubmitAdminProposal, sdk.NewAttribute(govtypes.AttributeKeyProposalType, msg.GetContent().ProposalType()))
	ctx.EventManager().EmitEvent(submitEvent)

	return &types.MsgSubmitProposalResponse{
		ProposalId: proposal.ProposalId,
	}, nil
}
