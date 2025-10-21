package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSON type alias for JSONB fields
type JSON map[string]interface{}

// Value marshals JSON for database storage
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan unmarshals JSON from database
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
	return json.Unmarshal(bytes, j)
}

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityGistCreated       ActivityType = "gist_created"
	ActivityGistUpdated       ActivityType = "gist_updated"
	ActivityGistDeleted       ActivityType = "gist_deleted"
	ActivityGistStarred       ActivityType = "gist_starred"
	ActivityGistUnstarred     ActivityType = "gist_unstarred"
	ActivityGistForked        ActivityType = "gist_forked"
	ActivityGistCommented     ActivityType = "gist_commented"
	ActivityUserFollowed      ActivityType = "user_followed"
	ActivityUserUnfollowed    ActivityType = "user_unfollowed"
	ActivityUserRegistered    ActivityType = "user_registered"
	ActivityOrgCreated        ActivityType = "org_created"
	ActivityOrgMemberAdded    ActivityType = "org_member_added"
	ActivityOrgMemberRemoved  ActivityType = "org_member_removed"
)

// ActivityFeed represents an activity in the system
type ActivityFeed struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActorID      uuid.UUID       `gorm:"type:uuid;not null" json:"actor_id"`
	Actor        *User           `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
	Type         ActivityType    `gorm:"type:varchar(50);not null" json:"type"`
	TargetType   string          `gorm:"type:varchar(50)" json:"target_type,omitempty"`
	TargetID     *uuid.UUID      `gorm:"type:uuid" json:"target_id,omitempty"`
	SecondaryID  *uuid.UUID      `gorm:"type:uuid" json:"secondary_id,omitempty"`
	Visibility   string          `gorm:"type:varchar(20);default:'public'" json:"visibility"`
	Metadata     JSON            `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// ActivityFeedFollow represents users following activity feeds
type ActivityFeedFollow struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FollowedID uuid.UUID  `gorm:"type:uuid;not null" json:"followed_id"`
	Followed   *User      `gorm:"foreignKey:FollowedID" json:"followed,omitempty"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// ActivityFeedSubscription represents subscription preferences
type ActivityFeedSubscription struct {
	ID         uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	User       *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Types      []string     `gorm:"type:text[]" json:"types"`
	EmailNotify bool        `gorm:"default:false" json:"email_notify"`
	WebNotify   bool        `gorm:"default:true" json:"web_notify"`
	CreatedAt   time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// CreateActivity creates a new activity feed entry
func CreateActivity(db *gorm.DB, activity *ActivityFeed) error {
	return db.Create(activity).Error
}

// GetUserFeed gets activity feed for a user (including followed users)
func GetUserFeed(db *gorm.DB, userID uuid.UUID, limit, offset int) ([]ActivityFeed, error) {
	var activities []ActivityFeed

	// Get list of users this user follows
	var followedIDs []uuid.UUID
	db.Model(&ActivityFeedFollow{}).
		Where("user_id = ?", userID).
		Pluck("followed_id", &followedIDs)

	// Include self in the list
	followedIDs = append(followedIDs, userID)

	// Get activities from followed users
	err := db.Preload("Actor").
		Where("actor_id IN (?) AND visibility = ?", followedIDs, "public").
		Or("actor_id = ?", userID). // Always show own activities
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	return activities, err
}

// GetPublicFeed gets public activity feed
func GetPublicFeed(db *gorm.DB, limit, offset int) ([]ActivityFeed, error) {
	var activities []ActivityFeed

	err := db.Preload("Actor").
		Where("visibility = ?", "public").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	return activities, err
}

// GetUserActivities gets activities performed by a specific user
func GetUserActivities(db *gorm.DB, userID uuid.UUID, limit, offset int) ([]ActivityFeed, error) {
	var activities []ActivityFeed

	err := db.Preload("Actor").
		Where("actor_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	return activities, err
}

// FollowUser follows another user's activity feed
func FollowUser(db *gorm.DB, userID, followedID uuid.UUID) error {
	follow := &ActivityFeedFollow{
		UserID:     userID,
		FollowedID: followedID,
	}

	// Check if already following
	var count int64
	db.Model(&ActivityFeedFollow{}).
		Where("user_id = ? AND followed_id = ?", userID, followedID).
		Count(&count)

	if count > 0 {
		return nil // Already following
	}

	// Create follow relationship
	if err := db.Create(follow).Error; err != nil {
		return err
	}

	// Create activity for following
	activity := &ActivityFeed{
		ActorID:     userID,
		Type:        ActivityUserFollowed,
		TargetType:  "user",
		TargetID:    &followedID,
		Visibility:  "public",
	}

	return CreateActivity(db, activity)
}

// UnfollowUser unfollows another user's activity feed
func UnfollowUser(db *gorm.DB, userID, followedID uuid.UUID) error {
	// Delete follow relationship
	if err := db.Where("user_id = ? AND followed_id = ?", userID, followedID).
		Delete(&ActivityFeedFollow{}).Error; err != nil {
		return err
	}

	// Create activity for unfollowing
	activity := &ActivityFeed{
		ActorID:     userID,
		Type:        ActivityUserUnfollowed,
		TargetType:  "user",
		TargetID:    &followedID,
		Visibility:  "public",
	}

	return CreateActivity(db, activity)
}

// GetFollowers gets followers of a user
func GetFollowers(db *gorm.DB, userID uuid.UUID, limit, offset int) ([]User, error) {
	var followers []User

	err := db.Table("users").
		Joins("JOIN activity_feed_follows ON users.id = activity_feed_follows.user_id").
		Where("activity_feed_follows.followed_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Find(&followers).Error

	return followers, err
}

// GetFollowing gets users that a user is following
func GetFollowing(db *gorm.DB, userID uuid.UUID, limit, offset int) ([]User, error) {
	var following []User

	err := db.Table("users").
		Joins("JOIN activity_feed_follows ON users.id = activity_feed_follows.followed_id").
		Where("activity_feed_follows.user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Find(&following).Error

	return following, err
}

// LogGistActivity logs gist-related activities
func LogGistActivity(db *gorm.DB, actorID uuid.UUID, activityType ActivityType, gistID uuid.UUID, visibility string) error {
	activity := &ActivityFeed{
		ActorID:    actorID,
		Type:       activityType,
		TargetType: "gist",
		TargetID:   &gistID,
		Visibility: visibility,
	}
	return CreateActivity(db, activity)
}

// CleanupOldActivities removes activities older than specified days
func CleanupOldActivities(db *gorm.DB, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return db.Where("created_at < ?", cutoff).Delete(&ActivityFeed{}).Error
}

// GetActivityStats gets statistics about activities
func GetActivityStats(db *gorm.DB, userID uuid.UUID, period time.Duration) (map[string]int64, error) {
	stats := make(map[string]int64)
	cutoff := time.Now().Add(-period)

	// Count activities by type
	rows, err := db.Model(&ActivityFeed{}).
		Select("type, COUNT(*) as count").
		Where("actor_id = ? AND created_at > ?", userID, cutoff).
		Group("type").
		Rows()

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var activityType string
		var count int64
		if err := rows.Scan(&activityType, &count); err != nil {
			continue
		}
		stats[activityType] = count
	}

	return stats, nil
}