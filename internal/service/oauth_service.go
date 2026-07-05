package service

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/pkg/logger"
	"backend/pkg/utils"
	"crypto/sha256"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	ProviderApple  = "apple"
	ProviderGoogle = "google"
	ProviderWechat = "wechat"
	ProviderAlipay = "alipay"
	ProviderEmail  = "email"
)

type OAuthLoginRequest struct {
	Provider    string                 `json:"provider"`
	Code        string                 `json:"code"`
	IDToken     string                 `json:"id_token"`
	AccessToken string                 `json:"access_token"`
	Email       string                 `json:"email"`
	Extra       map[string]interface{} `json:"extra"`
	IP          string
	UA          string
	DeviceID    string
}

type OAuthLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	NewUser      bool   `json:"new_user"`
}

var validProviders = map[string]bool{
	ProviderApple:  true,
	ProviderGoogle: true,
	ProviderWechat: true,
	ProviderAlipay: true,
	ProviderEmail:  true,
}

func isValidProvider(provider string) bool {
	return validProviders[provider]
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(pattern).MatchString(email)
}

type OAuthProfile struct {
	ProviderID  string
	Email       string
	Name        string
	AvatarURL   string
	Extra       map[string]interface{}
}

func mockValidateAppleToken(idToken string) (*OAuthProfile, error) {
	return &OAuthProfile{
		ProviderID:  "mock_apple_user_" + idToken[:8],
		Email:       "user_" + idToken[:8] + "@example.com",
		Name:        "Apple User",
		AvatarURL:   "",
		Extra:       map[string]interface{}{"id_token": idToken},
	}, nil
}

func mockValidateGoogleToken(idToken string) (*OAuthProfile, error) {
	return &OAuthProfile{
		ProviderID:  "mock_google_user_" + idToken[:8],
		Email:       "user_" + idToken[:8] + "@gmail.com",
		Name:        "Google User",
		AvatarURL:   "https://lh3.googleusercontent.com/mock",
		Extra:       map[string]interface{}{"id_token": idToken},
	}, nil
}

func mockValidateWechatCode(code string) (*OAuthProfile, error) {
	return &OAuthProfile{
		ProviderID:  "mock_wechat_openid_" + code[:8],
		Email:       "",
		Name:        "微信用户",
		AvatarURL:   "https://wx.qlogo.cn/mock",
		Extra:       map[string]interface{}{"code": code, "openid": "mock_openid_" + code[:8]},
	}, nil
}

func mockValidateAlipayCode(code string) (*OAuthProfile, error) {
	return &OAuthProfile{
		ProviderID:  "mock_alipay_userid_" + code[:8],
		Email:       "",
		Name:        "支付宝用户",
		AvatarURL:   "https://mobile.alipay.com/mock",
		Extra:       map[string]interface{}{"code": code, "user_id": "mock_userid_" + code[:8]},
	}, nil
}

func validateEmailLogin(email string) (*OAuthProfile, error) {
	if !isValidEmail(email) {
		return nil, fmt.Errorf("邮箱格式不正确")
	}
	return &OAuthProfile{
		ProviderID:  email,
		Email:       email,
		Name:        "",
		AvatarURL:   "",
		Extra:       map[string]interface{}{"email": email},
	}, nil
}

func validateOAuthToken(req OAuthLoginRequest) (*OAuthProfile, error) {
	switch req.Provider {
	case ProviderApple:
		if req.IDToken == "" {
			return nil, fmt.Errorf("Apple登录需要id_token")
		}
		return mockValidateAppleToken(req.IDToken)
	case ProviderGoogle:
		if req.IDToken == "" {
			return nil, fmt.Errorf("Google登录需要id_token")
		}
		return mockValidateGoogleToken(req.IDToken)
	case ProviderWechat:
		if req.Code == "" {
			return nil, fmt.Errorf("微信登录需要code")
		}
		return mockValidateWechatCode(req.Code)
	case ProviderAlipay:
		if req.Code == "" {
			return nil, fmt.Errorf("支付宝登录需要code")
		}
		return mockValidateAlipayCode(req.Code)
	case ProviderEmail:
		if req.Email == "" {
			return nil, fmt.Errorf("邮箱登录需要email")
		}
		return validateEmailLogin(req.Email)
	default:
		return nil, fmt.Errorf("不支持的登录方式")
	}
}

func OAuthLogin(req OAuthLoginRequest) (*OAuthLoginResponse, error) {
	if !isValidProvider(req.Provider) {
		return nil, fmt.Errorf("不支持的登录方式: %s", req.Provider)
	}

	logger.Infof("[OAuth] OAuthLogin start - Provider: %s, IP: %s", req.Provider, req.IP)

	profile, err := validateOAuthToken(req)
	if err != nil {
		logger.Warnf("[OAuth] OAuthLogin failed - Provider: %s, Error: %v", req.Provider, err)
		return nil, err
	}

	oauthUser, err := repository.GetOAuthUser(req.Provider, profile.ProviderID)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("验证失败")
	}

	var user *model.User
	newUser := false

	if oauthUser != nil {
		user, err = repository.GetUserByID(oauthUser.UserID)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
	} else {
		if profile.Email != "" {
			existingUser, err := repository.GetUserByEmail(profile.Email)
			if err == nil && existingUser.ID != 0 {
				user = existingUser
				err = repository.CreateOAuthUser(existingUser.ID, req.Provider, profile.ProviderID, profile.Extra)
				if err != nil {
					return nil, fmt.Errorf("关联账号失败")
				}
			}
		}

		if user == nil {
			newUser = true
			password := generateRandomCode(8)
			passwordHash, _ := generatePasswordHash(password)

			phoneNum := generateVirtualPhone(req.Provider, profile.ProviderID)

			user = &model.User{
				Uid:           utils.GenerateUID(),
				PhoneNum:      phoneNum,
				Password:      passwordHash,
				PlainPassword: password,
				Nickname:      profile.Name,
				AvatarURL:     profile.AvatarURL,
				Email:         profile.Email,
				Gender:        -1,
				AccountStatus: 1,
			}

			err = repository.CreateUser(user)
			if err != nil {
				logger.Errorf("[OAuth] CreateUser failed - Error: %v", err)
				return nil, fmt.Errorf("注册失败")
			}

			err = repository.CreateOAuthUser(user.ID, req.Provider, profile.ProviderID, profile.Extra)
			if err != nil {
				logger.Errorf("[OAuth] CreateOAuthUser failed - Error: %v", err)
				return nil, fmt.Errorf("关联账号失败")
			}
		}
	}

	if user.AccountStatus == 0 {
		return nil, fmt.Errorf("账号已被封禁")
	}
	if user.AccountStatus == 2 {
		return nil, fmt.Errorf("账号已注销")
	}

	accessToken, err := generateToken(user.Uid)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败")
	}

	refreshToken, err := generateRefreshToken(user.Uid)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败")
	}

	repository.SaveUserToken(user.Uid, accessToken)
	repository.SaveRefreshToken(user.Uid, refreshToken)

	repository.LogOperation(user.ID, req.IP, req.UA, "login_oauth_"+req.Provider, true, "第三方登录成功")

	logger.Infof("[OAuth] OAuthLogin success - Provider: %s, UserID: %d, NewUser: %v", req.Provider, user.ID, newUser)

	return &OAuthLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		NewUser:      newUser,
	}, nil
}

func generateVirtualPhone(provider, providerID string) string {
	hash := generateNumericHash(provider + ":" + providerID)
	return "9" + hash[:9]
}

func generateNumericHash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	sum := h.Sum(nil)
	
	var result string
	for i := 0; i < len(sum) && len(result) < 10; i++ {
		digit := int(sum[i] % 10)
		result += fmt.Sprintf("%d", digit)
	}
	return result
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func generateToken(uid string) (string, error) {
	return middleware.GenerateToken(uid)
}

func generateRefreshToken(uid string) (string, error) {
	return middleware.GenerateRefreshToken(uid)
}

type OAuthBindRequest struct {
	UID         string                 `json:"uid"`
	Provider    string                 `json:"provider"`
	Code        string                 `json:"code"`
	IDToken     string                 `json:"id_token"`
	AccessToken string                 `json:"access_token"`
	Email       string                 `json:"email"`
	Extra       map[string]interface{} `json:"extra"`
	IP          string
	UA          string
}

func OAuthBind(req OAuthBindRequest) error {
	if !isValidProvider(req.Provider) {
		return fmt.Errorf("不支持的绑定方式: %s", req.Provider)
	}

	user, err := repository.GetUserByUID(req.UID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	profile, err := validateOAuthToken(OAuthLoginRequest{
		Provider:    req.Provider,
		Code:        req.Code,
		IDToken:     req.IDToken,
		AccessToken: req.AccessToken,
		Email:       req.Email,
		Extra:       req.Extra,
	})
	if err != nil {
		return err
	}

	existingOAuth, err := repository.GetOAuthUser(req.Provider, profile.ProviderID)
	if err == nil && existingOAuth != nil {
		if existingOAuth.UserID != user.ID {
			return fmt.Errorf("该账号已绑定其他用户")
		}
		return nil
	}

	err = repository.CreateOAuthUser(user.ID, req.Provider, profile.ProviderID, profile.Extra)
	if err != nil {
		return fmt.Errorf("绑定失败")
	}

	repository.LogOperation(user.ID, req.IP, req.UA, "oauth_bind_"+req.Provider, true, "绑定成功")

	return nil
}

type OAuthUnbindRequest struct {
	UID      string
	Provider string
	IP       string
	UA       string
}

func OAuthUnbind(req OAuthUnbindRequest) error {
	if !isValidProvider(req.Provider) {
		return fmt.Errorf("不支持的解绑方式: %s", req.Provider)
	}

	user, err := repository.GetUserByUID(req.UID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	err = repository.DeleteOAuthUser(user.ID, req.Provider)
	if err != nil {
		return fmt.Errorf("解绑失败")
	}

	repository.LogOperation(user.ID, req.IP, req.UA, "oauth_unbind_"+req.Provider, true, "解绑成功")

	return nil
}

type OAuthListRequest struct {
	UID string
}

type OAuthListResponse struct {
	Providers []string `json:"providers"`
}

func GetOAuthBindings(req OAuthListRequest) (*OAuthListResponse, error) {
	user, err := repository.GetUserByUID(req.UID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	oauthUsers, err := repository.GetOAuthUsersByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}

	providers := make([]string, 0)
	for _, oauth := range oauthUsers {
		providers = append(providers, oauth.Provider)
	}

	return &OAuthListResponse{
		Providers: providers,
	}, nil
}