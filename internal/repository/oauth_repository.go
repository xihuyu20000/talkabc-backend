package repository

import (
	"backend/internal/config"
	"backend/internal/model"
	"errors"
)

func GetOAuthUser(provider, providerID string) (*model.OAuthUser, error) {
	var oauthUser model.OAuthUser
	err := config.DB.Where("provider = ? AND provider_id = ?", provider, providerID).First(&oauthUser).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, err
		}
		return nil, err
	}
	return &oauthUser, nil
}

func CreateOAuthUser(userID uint, provider, providerID string, extra map[string]interface{}) error {
	oauthUser := &model.OAuthUser{
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
		Extra:      extra,
	}
	return config.DB.Create(oauthUser).Error
}

func UpdateOAuthUser(userID uint, provider string, extra map[string]interface{}) error {
	return config.DB.Model(&model.OAuthUser{}).
		Where("user_id = ? AND provider = ?", userID, provider).
		Update("extra", extra).Error
}

func DeleteOAuthUser(userID uint, provider string) error {
	return config.DB.Where("user_id = ? AND provider = ?", userID, provider).Delete(&model.OAuthUser{}).Error
}

func GetOAuthUsersByUserID(userID uint) ([]model.OAuthUser, error) {
	var oauthUsers []model.OAuthUser
	err := config.DB.Where("user_id = ?", userID).Find(&oauthUsers).Error
	if err != nil {
		return nil, err
	}
	return oauthUsers, nil
}

func GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, errors.New("record not found")
		}
		return nil, err
	}
	return &user, nil
}