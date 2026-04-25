package usersvc

import (
	"context"
	"testing"
)

func TestTestReservedWord(t *testing.T) {
	t.Parallel()

	service := userService{
		reservedWords: &[]string{"admin", "settings"},
	}

	if !service.TestReservedWord("admin") {
		t.Fatal("expected admin to be reserved")
	}
	if service.TestReservedWord("chef-user") {
		t.Fatal("did not expect chef-user to be reserved")
	}
}

func TestTestValidName(t *testing.T) {
	t.Parallel()

	service := userService{}

	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "minimum valid length", input: "chefuser", want: true},
		{name: "hyphenated", input: "chef-user", want: true},
		{name: "too short", input: "short", want: false},
		{name: "starts with hyphen", input: "-chefuser", want: false},
		{name: "ends with hyphen", input: "chefuser-", want: false},
		{name: "double hyphen", input: "chef--user", want: false},
		{name: "invalid character", input: "chef_user", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := service.TestValidName(tc.input); got != tc.want {
				t.Fatalf("TestValidName(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestTestName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		input     string
		reserved  []string
		available func(context.Context, string) (bool, error)
		wantErr   error
	}{
		{
			name:     "invalid syntax",
			input:    "bad",
			reserved: nil,
			wantErr:  ErrInvalidUsername,
		},
		{
			name:     "reserved word",
			input:    "adminuser",
			reserved: []string{"adminuser"},
			wantErr:  ErrReservedWord,
		},
		{
			name:     "username already in use",
			input:    "chef-user",
			reserved: nil,
			available: func(context.Context, string) (bool, error) {
				return true, ErrUsernameInUse
			},
			wantErr: ErrUsernameInUse,
		},
		{
			name:     "available username",
			input:    "chef-user",
			reserved: nil,
			available: func(context.Context, string) (bool, error) {
				return false, nil
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := userService{
				reservedWords: &tc.reserved,
			}
			if tc.available != nil {
				service.userCollection = nil
				original := service.TestAvailableName
				_ = original
			}

			if tc.available != nil {
				service := testableUserService{
					userService:         service,
					testAvailableNameFn: tc.available,
				}
				if err := service.TestName(context.Background(), tc.input); err != tc.wantErr {
					t.Fatalf("TestName(%q) error = %v, want %v", tc.input, err, tc.wantErr)
				}
				return
			}

			if err := service.TestName(context.Background(), tc.input); err != tc.wantErr {
				t.Fatalf("TestName(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

type testableUserService struct {
	userService
	testAvailableNameFn func(context.Context, string) (bool, error)
}

func (s testableUserService) TestAvailableName(ctx context.Context, username string) (bool, error) {
	return s.testAvailableNameFn(ctx, username)
}

func (s testableUserService) TestName(ctx context.Context, n string) error {
	if isValid := s.TestValidName(n); !isValid {
		return ErrInvalidUsername
	}
	if isReserved := s.TestReservedWord(n); isReserved {
		return ErrReservedWord
	}
	if _, err := s.TestAvailableName(ctx, n); err != nil {
		if err == ErrUsernameInUse {
			return ErrUsernameInUse
		}
		if err != ErrUserNotFound {
			return err
		}
	}
	return nil
}
