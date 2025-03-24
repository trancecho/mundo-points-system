package domain

import (
	"context"
	"github.com/trancecho/mundo-points-system/pkg/meta"
	"github.com/trancecho/mundo-points-system/pkg/utils"
	"github.com/trancecho/mundo-points-system/po"
	v1 "github.com/trancecho/mundo-points-system/proto/point/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

const LikePoints = int64(1)

type UserService struct {
	v1.UnimplementedUserServiceServer
	userRepo  po.UserRepository
	pointRepo po.PointRepository
	statRepo  po.StatisticsRepository
}

func NewUserService(userRepo po.UserRepository, pointRepo po.PointRepository, statRepo po.StatisticsRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		pointRepo: pointRepo,
		statRepo:  statRepo,
	}
}

func (s *UserService) UpdatePointsAndExperience(ctx context.Context, req *v1.UpdatePointsRequest) (*v1.CommonResponse, error) {
	_, err := meta.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	userClaims, ok := ctx.Value("claims").(*utils.Claims)
	if !ok {
		return nil, status.Errorf(400, "failed to get user claims from context")
	}
	//如果扣除积分，需要查看用户积分是否足够
	if req.DeltaPoints < 0 {
		user, err := s.userRepo.GetUserByID(ctx, strconv.FormatInt(userClaims.UserID, 10))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "获取用户ID失败: %v", err)
		}

		if user.Points < -req.DeltaPoints {
			return &v1.CommonResponse{
				Success:   false,
				Message:   "积分不足",
				ErrorCode: v1.ErrorCode_POINTS_INSUFFICIENT,
			}, nil
		}
	}
	//更新积分和经验
	err = s.pointRepo.AddPointsAndExperience(ctx, strconv.FormatInt(userClaims.UserID, 10), req.DeltaPoints, req.DeltaExperience, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新积分和经验失败: %v", err)
	}
	//如果有经验变更，则可能需要更新等级
	if req.DeltaExperience != 0 {
		err = s.userRepo.UpdateLevelByExperience(ctx, strconv.FormatInt(userClaims.UserID, 10))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "更新等级失败: %v", err)
		}
	}

	return &v1.CommonResponse{
		Success:   true,
		Message:   "更新积分和经验成功",
		ErrorCode: v1.ErrorCode_NONE_ERROR,
	}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.UserInfo, error) {
	_, err := meta.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取用户信息失败: %v", err)
	}
	return &v1.UserInfo{
		UserId:             req.UserId,
		Username:           user.Username,
		Points:             user.Points,
		Experience:         user.Experience,
		Level:              int32(user.Level),
		ContinuousSignDays: user.ContinuousSignDay,
		TotalSignDays:      user.TotalSignDay,
	}, nil
}

func (s *UserService) ProcessLike(ctx context.Context, req *v1.LikeRequest) (*v1.CommonResponse, error) {
	_, err := meta.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	userClaims, ok := ctx.Value("userClaims").(*po.UserInfo)
	if !ok {
		return nil, status.Errorf(400, "failed to get user claims from context")
	}
	//记录点赞信息
	err = s.pointRepo.RecordLike(ctx, strconv.FormatInt(userClaims.UserID, 10), req.PostId, req.TargetUserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "点赞失败: %v", err)
	}

	return &v1.CommonResponse{
		Success:   true,
		Message:   "点赞成功",
		ErrorCode: v1.ErrorCode_NONE_ERROR,
	}, nil
}

// GetAdminStats 实现获取管理员统计数据功能
func (s *UserService) GetAdminStats(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.AdminStats, error) {
	_, err := meta.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	userClaims, ok := ctx.Value("claims").(*utils.Claims)
	if !ok {
		return nil, status.Errorf(400, "failed to get user claims from context")
	}
	//验证管理员权限
	if userClaims.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "用户无权限")
	}
	// 获取等级分布
	levelDistribution, err := s.statRepo.GetLevelDistribution(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取等级分布失败: %v", err)
	}

	// 获取平均积分
	avgPoints, err := s.statRepo.GetAveragePoints(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取平均积分失败: %v", err)
	}

	// 获取本月使用的积分
	monthlyPointsUsed, err := s.statRepo.GetMonthlyPointsUsed(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取本月使用积分失败: %v", err)
	}

	// 构造返回数据
	stats := &v1.AdminStats{
		AvgPoints:         avgPoints,
		MonthlyPointsUsed: monthlyPointsUsed,
		LevelDistribution: make([]*v1.LevelDistribution, 0, len(levelDistribution)),
	}

	for level, count := range levelDistribution {
		stats.LevelDistribution = append(stats.LevelDistribution, &v1.LevelDistribution{
			Level:     int32(level),
			UserCount: count,
		})
	}

	return stats, nil
}
