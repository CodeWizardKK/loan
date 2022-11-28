package keeper

import (
	"context"
	"fmt"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RepayLoan(goCtx context.Context, msg *types.MsgRepayLoan) (*types.MsgRepayLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//特定のloanを取得する
	loan, isFound := k.GetLoan(ctx, msg.Id)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	//特定のloanのステータスが承認済みか確認する
	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "%v", loan.State)
	}

	//ローンを返済できるのは本人だけ
	if loan.Borrower != msg.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Cannot repay: not the borrower")
	}

	// //型変換(返済期限)
	// deadline, _ := strconv.ParseInt(loan.Deadline, 10, 64)

	// //返済期限が過ぎている場合は、終了
	// if deadline < ctx.BlockHeight() {
	// 	return nil, sdkerrors.Wrap(types.ErrDeadline, "Cannot repay before deadline")
	// }

	//型変換
	lender, _ := sdk.AccAddressFromBech32(loan.Lender)
	borrower, _ := sdk.AccAddressFromBech32(loan.Borrower)
	amount, _ := sdk.ParseCoinsNormalized(loan.Amount)
	fee, _ := sdk.ParseCoinsNormalized(loan.Fee)
	collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)

	//借り手から貸し手へローン金額(amount)、ローン手数料(fee)を渡す
	err := k.bankKeeper.SendCoins(ctx, borrower, lender, amount)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Cannot send coins")
	}
	err = k.bankKeeper.SendCoins(ctx, borrower, lender, fee)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Cannot send coins")
	}

	//モジュールアカウント(エスクローアカウントとして使用)から借り手へ担保(collateral)を返す。
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrower, collateral)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Cannot send coins")
	}

	//ローンが返済されたので、ローン情報を更新する
	loan.State = "paid"

	//loanストア更新
	k.SetLoan(ctx, loan)

	return &types.MsgRepayLoanResponse{}, nil
}
