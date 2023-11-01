package adminmodule

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"admin-module/x/adminmodule/keeper"
	"admin-module/x/adminmodule/types"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := keeper.Logger(ctx)

	keeper.IterateActiveProposalsQueue(ctx, func(proposal govtypes.Proposal) bool {
		var logMsg, tagValue string

		handler := keeper.Router().GetRoute(proposal.ProposalRoute())
		cacheCtx, writeCache := ctx.CacheContext()

		// The proposal handler may execute state mutating logic depending
		// on the proposal content. If the handler fails, no state mutation
		// is written and the error message is logged.
		err := handler(cacheCtx, proposal.GetContent())
		if err == nil {
			logMsg = "passed"
			proposal.Status = govtypes.StatusPassed
			tagValue = gov.AttributeValueProposalPassed

			// The cached context is created with a new EventManager. However, since
			// the proposal handler execution was successful, we want to track/keep
			// any events emitted, so we re-emit to "merge" the events into the
			// original Context's EventManager.
			ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

			// write state to the underlying multi-store
			writeCache()
		} else {
			proposal.Status = govtypes.StatusFailed
			tagValue = gov.AttributeValueProposalFailed
			logMsg = fmt.Sprintf("proposal failed on execution: %s", err)
		}

		keeper.SetProposal(ctx, proposal)
		keeper.RemoveFromActiveProposalQueue(ctx, proposal.ProposalId)

		keeper.AddToArchive(ctx, proposal)

		logger.Info(
			"proposal tallied",
			"proposal", proposal.ProposalId,
			"title", proposal.GetTitle(),
			"result", logMsg,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAdminProposal,
				sdk.NewAttribute(gov.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ProposalId)),
				sdk.NewAttribute(gov.AttributeKeyProposalResult, tagValue),
			),
		)
		return false
	})
}
