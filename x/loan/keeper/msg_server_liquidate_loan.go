package keeper

import (
	"context"
	"strconv"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) LiquidateLoan(goCtx context.Context, msg *types.MsgLiquidateLoan) (*types.MsgLiquidateLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//特定のloanを取得する
	loan, isFound := k.GetLoan(ctx, msg.Id)
	if !isFound {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "key %d doesn't exist", msg.Id)
	}

	//
	if loan.Lender != msg.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Cannot liquidate: not the lender")
	}

	//特定のloanのステータスが承認済みか確認する
	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "%v", loan.State)
	}

	//型変換
	lender, _ := sdk.AccAddressFromBech32(loan.Lender)
	collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)

	//型変換(返済期限)
	deadline, err := strconv.ParseInt(loan.Deadline, 10, 64)
	if err != nil {
		panic(err)
	}

	//返済期限がまだ過ぎていない場合は、終了
	if ctx.BlockHeight() < deadline {
		return nil, sdkerrors.Wrap(types.ErrDeadline, "Cannot liquidate before deadline")
	}

	//モジュールアカウントから貸し手に担保を送る
	k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lender, collateral)

	//ローンが清算されたので、ローン情報を更新する
	loan.State = "liquidated"

	//loanストア更新
	k.SetLoan(ctx, loan)

	return &types.MsgLiquidateLoanResponse{}, nil
}
