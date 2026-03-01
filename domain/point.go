package domain

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/trancecho/mundo-points-system/pkg/meta"
	"github.com/trancecho/mundo-points-system/pkg/utils"
	"github.com/trancecho/mundo-points-system/po"
	v1 "github.com/trancecho/mundo-points-system/proto/point/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		ActivityScore:      user.ActivityScore,
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

func (s *UserService) Sign(ctx context.Context, req *v1.SignRequest) (*v1.CommonResponse, error) {
	_, err := meta.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	//获取用户信息
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取用户信息失败: %v", err)
	}
	//获取当前时间
	nowtime := time.Now()
	//检查用户是否在同一天签到
	if !user.LastSignDate.IsZero() {
		lastSignYear, lastSignMonth, lastSignDay := user.LastSignDate.Date()
		currentYear, currentMonth, currentDay := nowtime.Date()

		// 如果在同一天已经签到
		if lastSignYear == currentYear && lastSignMonth == currentMonth && lastSignDay == currentDay {
			return &v1.CommonResponse{
				Success:   false,
				Message:   "今日已签到",
				ErrorCode: v1.ErrorCode_INVALID_REQUEST,
			}, nil
		}
		//检查是否连续签到
		yesterday := nowtime.AddDate(0, 0, -1)
		yesterdayYear, yesterdayMonth, yesterdayDay := yesterday.Date()
		// 如果上次签到不是昨天，连续签到天数重置为1
		if !(lastSignYear == yesterdayYear && lastSignMonth == yesterdayMonth && lastSignDay == yesterdayDay) {
			user.ContinuousSignDay = 0 // 重置连续签到次数
		}
	}
	// 计算签到奖励
	pointsReward := int64(50)  // 基础积分奖励
	expReward := int64(10)     // 基础经验奖励
	activityReward := int64(5) // 基础活跃度奖励

	// 连续签到奖励加成
	continuousDay := user.ContinuousSignDay + 1

	// 根据活跃度计算加成
	activityBonus := 0.0

	// 直接使用 user.ActivityScore
	if user.ActivityScore >= 10000 {
		activityBonus = 0.5
	} else if user.ActivityScore >= 5000 {
		activityBonus = 0.3
	} else if user.ActivityScore >= 2000 {
		activityBonus = 0.2
	} else if user.ActivityScore >= 500 {
		activityBonus = 0.1
	}

	// 应用活跃度加成
	pointsReward = pointsReward + int64(float64(pointsReward)*activityBonus)
	expReward = expReward + int64(float64(expReward)*activityBonus)

	// 更新用户签到信息
	totalDay := user.TotalSignDay + 1

	err = s.userRepo.UpdateSignStatus(ctx, req.UserId, true, continuousDay, totalDay)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新签到状态失败: %v", err)
	}

	// 添加积分和经验
	err = s.pointRepo.AddPointsAndExperience(ctx, req.UserId, pointsReward, expReward, "每日签到")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "添加积分和经验失败: %v", err)
	}

	// 直接更新用户活跃度
	err = s.userRepo.UpdateActivityScore(ctx, req.UserId, activityReward)
	if err != nil {
		// 只记录日志，不影响签到主流程
		log.Printf("更新用户活跃度失败: %v", err)
	}

	// 返回签到成功及奖励信息
	return &v1.CommonResponse{
		Success:   true,
		Message:   fmt.Sprintf("签到成功，获得积分: %d, 经验: %d, 活跃度: %d", pointsReward, expReward, activityReward),
		ErrorCode: v1.ErrorCode_NONE_ERROR,
	}, nil
}
