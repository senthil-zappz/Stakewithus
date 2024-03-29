package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// HasReport checks if the report of this ID triple exists in the storage.
func (k Keeper) HasReport(ctx sdk.Context, rid types.RequestID, val sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has(types.ReportsOfValidatorPrefixKey(rid, val))
}

// SetDataReport saves the report to the storage without performing validation.
func (k Keeper) SetReport(ctx sdk.Context, rid types.RequestID, rep types.Report) {
	val, _ := sdk.ValAddressFromBech32(rep.Validator)
	key := types.ReportsOfValidatorPrefixKey(rid, val)
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshal(&rep))
}

// AddReports performs sanity checks and adds a new batch from one validator to one request
// to the store. Note that we expect each validator to report to all raw data requests at once.
func (k Keeper) AddReport(ctx sdk.Context, rid types.RequestID, rep types.Report) error {
	if err := k.CheckValidReport(ctx, rid, rep); err != nil {
		return err
	}
	k.SetReport(ctx, rid, rep)
	return nil
}

func (k Keeper) CheckValidReport(ctx sdk.Context, rid types.RequestID, rep types.Report) error {
	req, err := k.GetRequest(ctx, rid)
	if err != nil {
		return err
	}
	val, err := sdk.ValAddressFromBech32(rep.Validator)
	if err != nil {
		return err
	}
	reqVals := make([]sdk.ValAddress, len(req.RequestedValidators))
	for idx, reqVal := range req.RequestedValidators {
		v, err := sdk.ValAddressFromBech32(reqVal)
		if err != nil {
			return err
		}
		reqVals[idx] = v
	}
	if !ContainsVal(reqVals, val) {
		return sdkerrors.Wrapf(
			types.ErrValidatorNotRequested, "reqID: %d, val: %s", rid, rep.Validator)
	}
	if k.HasReport(ctx, rid, val) {
		return sdkerrors.Wrapf(
			types.ErrValidatorAlreadyReported, "reqID: %d, val: %s", rid, rep.Validator)
	}
	if len(rep.RawReports) != len(req.RawRequests) {
		return types.ErrInvalidReportSize
	}
	for _, rep := range rep.RawReports {
		// Here we can safely assume that external IDs are unique, as this has already been
		// checked by ValidateBasic performed in baseapp's runTx function.
		if !ContainsEID(req.RawRequests, rep.ExternalID) {
			return sdkerrors.Wrapf(
				types.ErrRawRequestNotFound, "reqID: %d, extID: %d", rid, rep.ExternalID)
		}
	}
	return nil
}

// GetReportIterator returns the iterator for all reports of the given request ID.
func (k Keeper) GetReportIterator(ctx sdk.Context, rid types.RequestID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ReportStoreKey(rid))
}

// GetReportCount returns the number of reports for the given request ID.
func (k Keeper) GetReportCount(ctx sdk.Context, rid types.RequestID) (count uint64) {
	iterator := k.GetReportIterator(ctx, rid)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// GetReports returns all reports for the given request ID, or nil if there is none.
func (k Keeper) GetReports(ctx sdk.Context, rid types.RequestID) (reports []types.Report) {
	iterator := k.GetReportIterator(ctx, rid)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var rep types.Report
		k.cdc.MustUnmarshal(iterator.Value(), &rep)
		reports = append(reports, rep)
	}
	return reports
}

// DeleteReports removes all reports for the given request ID.
func (k Keeper) DeleteReports(ctx sdk.Context, rid types.RequestID) {
	var keys [][]byte
	iterator := k.GetReportIterator(ctx, rid)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
	}
	for _, key := range keys {
		ctx.KVStore(k.storeKey).Delete(key)
	}
}
