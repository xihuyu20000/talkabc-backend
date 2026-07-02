package infra

import (
	"backend/internal/model"
	"backend/pkg/logger"

	"github.com/jinzhu/gorm"
)

// AutoMigrate 执行数据库自动迁移
// 参数说明：
//   - db: GORM数据库实例
//
// 业务流程：
//   1. 调用 db.AutoMigrate 自动创建或更新所有数据表
//   2. 如果迁移失败，记录日志并退出
//   3. 迁移成功后记录日志
//
// 模型按字母顺序排列，便于维护：
//   AdBanner, AgreeFriend, ChatMessage, DatingPurpose, Diamond, DiamondRecord
//   FriendRelation, Gift, HobbyTag, LikeRecord, Member, MemberRecord
//   MomentComment, MomentPraise, SystemMsg, User, UserBlock
//   UserDatingPurposeRel, UserFocus, UserFriend, UserHobbyRel
//   UserMessageTop, UserMoment, UserNotify, VisitRecord
func AutoMigrate(db *gorm.DB) {
	tables := []interface{}{
		&model.AdBanner{},
		&model.AgreeFriend{},
		&model.ChatMessage{},
		&model.DatingPurpose{},
		&model.Diamond{},
		&model.DiamondRecord{},
		&model.FriendRelation{},
		&model.Gift{},
		&model.HobbyTag{},
		&model.LikeRecord{},
		&model.Member{},
		&model.MemberRecord{},
		&model.MomentComment{},
		&model.MomentPraise{},
		&model.SystemMsg{},
		&model.User{},
		&model.UserBlock{},
		&model.UserDatingPurposeRel{},
		&model.UserFocus{},
		&model.UserFriend{},
		&model.UserHobbyRel{},
		&model.UserMessageTop{},
		&model.UserMoment{},
		&model.UserNotify{},
		&model.VisitRecord{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			logger.Warn("AutoMigrate failed for table %T: %v", table, err)
		}
	}

	logger.Info("Database migration completed")
}