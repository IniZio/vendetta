package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractUserInfo(t *testing.T) {
	tests := []struct {
		name      string
		ghCLIPath string
		wantErr   bool
		validate  func(*UserInfo)
	}{
		{
			name:      "extract valid user info",
			ghCLIPath: "gh",
			wantErr:   false,
			validate: func(ui *UserInfo) {
				if ui != nil {
					assert.NotEmpty(t, ui.Username)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInfo, err := ExtractUserInfo(tt.ghCLIPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, userInfo)
				if tt.validate != nil {
					tt.validate(userInfo)
				}
			}
		})
	}
}

func TestUserInfoValidate(t *testing.T) {
	tests := []struct {
		name     string
		userInfo *UserInfo
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid user info",
			userInfo: &UserInfo{
				Username:  "testuser",
				UserID:    12345,
				AvatarURL: "https://avatars.githubusercontent.com/u/12345",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			userInfo: &UserInfo{
				UserID:    12345,
				AvatarURL: "https://avatars.githubusercontent.com/u/12345",
			},
			wantErr: true,
			errMsg:  "username",
		},
		{
			name: "zero user id",
			userInfo: &UserInfo{
				Username:  "testuser",
				AvatarURL: "https://avatars.githubusercontent.com/u/12345",
			},
			wantErr: true,
			errMsg:  "user_id",
		},
		{
			name: "missing avatar url",
			userInfo: &UserInfo{
				Username: "testuser",
				UserID:   12345,
			},
			wantErr: true,
			errMsg:  "avatar_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.userInfo.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeTempAndPersistentUserInfo(t *testing.T) {
	tests := []struct {
		name       string
		persistent *UserInfo
		temp       *UserInfo
		want       *UserInfo
	}{
		{
			name: "merge overlapping user info",
			persistent: &UserInfo{
				Username: "olduser",
				UserID:   123,
			},
			temp: &UserInfo{
				Username:  "newuser",
				UserID:    456,
				AvatarURL: "https://example.com/avatar.jpg",
			},
			want: &UserInfo{
				Username:  "newuser",
				UserID:    456,
				AvatarURL: "https://example.com/avatar.jpg",
			},
		},
		{
			name: "preserve persistent fields not in temp",
			persistent: &UserInfo{
				Username:  "user",
				UserID:    123,
				AvatarURL: "https://example.com/avatar.jpg",
			},
			temp: &UserInfo{
				Username: "user",
			},
			want: &UserInfo{
				Username:  "user",
				UserID:    123,
				AvatarURL: "https://example.com/avatar.jpg",
			},
		},
		{
			name:       "nil persistent returns temp",
			persistent: nil,
			temp: &UserInfo{
				Username: "user",
				UserID:   123,
			},
			want: &UserInfo{
				Username: "user",
				UserID:   123,
			},
		},
		{
			name:       "nil temp returns persistent",
			persistent: &UserInfo{Username: "user", UserID: 123},
			temp:       nil,
			want: &UserInfo{
				Username: "user",
				UserID:   123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := MergeTempAndPersistentUserInfo(tt.persistent, tt.temp)
			if merged != nil {
				assert.Equal(t, tt.want.Username, merged.Username)
				assert.Equal(t, tt.want.UserID, merged.UserID)
				assert.Equal(t, tt.want.AvatarURL, merged.AvatarURL)
			}
		})
	}
}

func TestParseUserInfoJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		wantErr  bool
		validate func(*UserInfo)
	}{
		{
			name:    "valid json",
			data:    `{"username":"testuser","user_id":123,"avatar_url":"https://example.com/avatar.jpg"}`,
			wantErr: false,
			validate: func(ui *UserInfo) {
				assert.Equal(t, "testuser", ui.Username)
				assert.Equal(t, int64(123), ui.UserID)
			},
		},
		{
			name:    "invalid json",
			data:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui, err := ParseUserInfoJSON([]byte(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(ui)
				}
			}
		})
	}
}
