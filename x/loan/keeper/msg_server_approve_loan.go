package keeper

import (
	"context"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ApproveLoan(goCtx context.Context, msg *types.MsgApproveLoan) (*types.MsgApproveLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//特定のloanを取得する
	loan, isFound := k.GetLoan(ctx, msg.Id)
	if !isFound {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "key %d doesn't exist", msg.Id)
	}

	//特定のloanのステータスが承認待ち(要求)か確認する
	if loan.State != "requested" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "%v", loan.State)
	}

	//型変換(貸し手)
	lender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	//型変換(借り手)
	borrower, err := sdk.AccAddressFromBech32(loan.Borrower)
	if err != nil {
		panic(err)
	}

	//型変換(ローン金額)
	amount, err := sdk.ParseCoinsNormalized(loan.Amount)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Cannot parse coins in loan amount")
	}

	//貸し手から借り手へローン金額(amount)が渡される
	sdkError := k.bankKeeper.SendCoins(ctx, lender, borrower, amount)
	if sdkError != nil {
		return nil, sdkError
	}

	//ローンが承認されたので、ローン情報を更新する
	loan.Lender = msg.Creator
	loan.State = "approved"

	//loanストア更新
	k.SetLoan(ctx, loan)

	return &types.MsgApproveLoanResponse{}, nil
}
