package service

import (
	"context"
	"mt4grpc/common"

	"mt4grpc/api_pb"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type UserService struct {
	log          *zap.Logger
	dealerLoader *common.DealerLoader
}

func NewUserService(dealerLoader *common.DealerLoader, log *zap.Logger) *UserService {
	return &UserService{dealerLoader: dealerLoader, log: log}
}

func (this *UserService) Add(_ context.Context, req *api_pb.AddUserRequest) (*api_pb.AddUserResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		this.log.Error(
			"Filed to load dealer",
			zap.String("token", req.Token),
			zap.String("name", req.User.Name),
			zap.String("email", req.User.Email),
		)
		return &api_pb.AddUserResponse{
			Message: err.Error(),
			Code:    0,
		}, err
	}

	user, err := dealer.CreateAccount(
		req.User.Name,
		req.User.Password,
		req.User.PasswordInvestor,
		req.User.Group,
		req.User.City,
		req.User.Email,
		req.User.Phone,
	)

	if err != nil {
		this.log.Info(
			"user added",
			zap.String("token", req.Token),
			zap.String("name", req.User.Name),
			zap.String("email", req.User.Email),
		)
		return &api_pb.AddUserResponse{
			Login:   int32(user.Login),
			Message: "Added",
			Code:    100,
		}, nil
	} else {
		this.log.Error(
			"Filed to add user",
			zap.String("token", req.Token),
			zap.String("name", req.User.Name),
			zap.String("email", req.User.Email),
		)
		return &api_pb.AddUserResponse{
			Message: err.Error(),
			Code:    0,
		}, err
	}
}

func (this *UserService) GetInfo(_ context.Context, req *api_pb.UserInfoRequest) (*api_pb.UserInfoResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		return nil, err
	}

	user, err := dealer.GetAsset(int(req.Login))

	if err == nil {
		u := &api_pb.UserInfo{
			Group:        user.Group,
			Name:         user.Name,
			Enabled:      int32(user.Enabled),
			Leverage:     int32(user.Leverage),
			Balance:      user.Balance,
			Credit:       user.Credit,
			AgentAccount: int32(user.AgentAccount),
		}
		return &api_pb.UserInfoResponse{
			User: u,
		}, nil
	} else {
		this.log.Error(
			"Filed to get user",
			zap.String("token", req.Token),
			zap.Int32("login", req.Login),
		)
		return &api_pb.UserInfoResponse{
			Message: err.Error(),
			Code:    0,
		}, err
	}

	return nil, nil
}

func (this *UserService) Update(_ context.Context, req *api_pb.UpdateUserRequest) (*api_pb.UpdateUserResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		this.log.Error(
			"Filed to load dealer",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
		)
		return &api_pb.UpdateUserResponse{
			Message: err.Error(),
			Code:    0,
		}, err
	}

	_, retErr := dealer.UpdateAccount(
		int(req.User.Login),
		req.User.Name,
		req.User.Password,
		req.User.PasswordInvestor,
		req.User.Group,
		req.User.City,
		req.User.Email,
		req.User.Phone,
	)

	if err == nil {
		this.log.Info(
			"user updated",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.String("name", req.User.Name),
			zap.String("email", req.User.Email),
		)
		return &api_pb.UpdateUserResponse{
			Message: "Updated",
			Code:    100,
		}, nil
	} else {
		this.log.Error(
			"Filed to update user",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.String("name", req.User.Name),
			zap.String("email", req.User.Email),
			zap.String("mt4_err", retErr.Error()),
		)
		return &api_pb.UpdateUserResponse{
			Message: err.Error(),
			Code:    0,
		}, retErr
	}
}

func (this *UserService) Delete(_ context.Context, req *api_pb.DeleteUserRequest) (*api_pb.DeleteUserResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		this.log.Error(
			"Filed to load dealer",
			zap.String("token", req.Token),
			zap.Int32("login", req.Login),
		)
		return &api_pb.DeleteUserResponse{
			Message: err.Error(),
			Code:    0,
		}, err
	}

	retErr := dealer.DeleteAccount(int(req.Login))

	if err == nil {
		this.log.Info(
			"user removed",
			zap.String("token", req.Token),
			zap.Int32("login", req.Login),
		)
		return &api_pb.DeleteUserResponse{
			Message: "Deleted",
			Code:    100,
		}, nil
	} else {
		this.log.Error(
			"Filed to delete user",
			zap.String("token", req.Token),
			zap.Int32("login", req.Login),
			zap.String("mt4_err", retErr.Error()),
		)
		return &api_pb.DeleteUserResponse{
			Message: err.Error(),
			Code:    0,
		}, retErr
	}
}

func (this *UserService) ResetPassword(_ context.Context, req *api_pb.ResetPasswordRequest) (*api_pb.ResetPasswordResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		this.log.Error(
			"Filed to load dealer",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.String("password", req.User.Password),
			zap.String("password_investor", req.User.PasswordInvestor),
			zap.String("mt4_err", err.Error()),
		)
		return &api_pb.ResetPasswordResponse{
			Message: err.Error(),
			Code:    0,
		}, err
	}

	var retErr error
	if len(req.User.Password) > 0 {
		retErr = dealer.ResetPassword(
			int(req.User.Login),
			req.User.Password,
			0,
		)
	} else {
		retErr = dealer.ResetPassword(
			int(req.User.Login),
			req.User.PasswordInvestor,
			1,
		)
	}

	if retErr == nil {
		this.log.Info(
			"user password reset",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
		)
		return &api_pb.ResetPasswordResponse{
			Message: "Reset",
			Code:    100,
		}, nil
	} else {
		this.log.Error(
			"Filed to reset password",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.String("password", req.User.Password),
			zap.String("password_investor", req.User.PasswordInvestor),
			zap.String("mt4_err", retErr.Error()),
		)
		return &api_pb.ResetPasswordResponse{
			Message: retErr.Error(),
			Code:    0,
		}, err
	}

	return nil, nil
}

func (this *UserService) Deposit(_ context.Context, req *api_pb.FundsRequest) (*api_pb.FundsResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		this.log.Error(
			"Filed to load dealer",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
		)
		return &api_pb.FundsResponse{}, err
	}

	var is_credit bool
	if req.User.IsCredit == 1 {
		is_credit = true
	}
	retErr := dealer.DepositAccount(
		int(req.User.Login),
		req.User.Amount,
		req.User.Comment,
		is_credit,
	)

	if retErr == nil {
		this.log.Info(
			"user deposited",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.Float64("amount", req.User.Amount),
		)
		user, err := dealer.GetAsset(int(req.User.Login))

		if err != nil {
			return &api_pb.FundsResponse{
				Balance: user.Balance,
				Credit:  user.Credit,
			}, nil
		} else {
			return &api_pb.FundsResponse{}, err
		}
	} else {
		this.log.Error(
			"Filed to deposit user",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.Float64("amount", req.User.Amount),
			zap.String("mt4_err", retErr.Error()),
		)
		return &api_pb.FundsResponse{}, retErr
	}

}

func (this *UserService) Withdraw(_ context.Context, req *api_pb.FundsRequest) (*api_pb.FundsResponse, error) {

	dealer, err := this.dealerLoader.Load(req.Token)

	if err != nil {
		this.log.Error(
			"Filed to load dealer",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
		)
		return &api_pb.FundsResponse{}, err
	}
	user, err := dealer.GetAsset(int(req.User.Login))

	if err != nil {
		this.log.Error(
			"Filed to load User",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.String("mt4_err", err.Error()),
		)
		return &api_pb.FundsResponse{}, err
	}

	if user.FreeMargin < req.User.Amount {
		this.log.Error(
			"Filed withdraw more than free margin",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.Float64("amount", req.User.Amount),
			zap.Float64("free_margin", user.FreeMargin),
		)
		return &api_pb.FundsResponse{Balance: user.Balance, Credit: user.Credit}, errors.New("Failed withdraw more than free margin")
	}

	var is_credit bool
	if req.User.IsCredit == 1 {
		is_credit = true
	}
	retErr := dealer.WithdrawAccount(
		int(req.User.Login),
		req.User.Amount,
		req.User.Comment,
		is_credit,
	)

	if retErr == nil {
		this.log.Info(
			"user withdrawan",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.Float64("amount", req.User.Amount),
		)

		return &api_pb.FundsResponse{
			Balance: user.Balance,
			Credit:  user.Credit,
		}, nil
	} else {
		this.log.Error(
			"Filed to withdraw user",
			zap.String("token", req.Token),
			zap.Int32("login", req.User.Login),
			zap.Float64("amount", req.User.Amount),
			zap.String("mt4_err", retErr.Error()),
		)
		return &api_pb.FundsResponse{
			Balance: user.Balance,
			Credit:  user.Credit,
		}, retErr
	}

}
