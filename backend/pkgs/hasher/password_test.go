package hasher

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()
	type args struct {
		password      string
		invalidInputs []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "letters_and_numbers",
			args: args{
				password:      "password123456788",
				invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
			},
		},
		{
			name: "letters_number_and_special",
			args: args{
				password:      "!2afj3214pofajip3142j;fa",
				invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
			},
		},
		{
			name: "extra_long_password",
			args: args{
				password:      "this_is_a_very_long_password_that_should_be_hashed_properly_and_still_work_with_the_check_function",
				invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
			},
		},
		{
			name: "empty_password",
			args: args{
				password:      "",
				invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			check, _ := CheckPasswordHash(tt.args.password, got)
			if !check {
				t.Errorf("CheckPasswordHash() failed to validate password=%v against hash=%v", tt.args.password, got)
			}

			for _, invalid := range tt.args.invalidInputs {
				check, _ := CheckPasswordHash(invalid, got)
				if check {
					t.Errorf("CheckPasswordHash() improperly validated password=%v against hash=%v", invalid, got)
				}
			}
		})
	}
}

func TestHashPasswordWithLegacyBcrypt(t *testing.T) {
	t.Parallel()
	type args struct {
		password      string
		invalidInputs []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "bcrypt_legacy_password",
			args: args{
				password:      "legacyPassword123",
				invalidInputs: []string{"wrongPassword", "123456", "anotherWrongPassword"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			legacyHash, _ := bcrypt.GenerateFromPassword([]byte(tt.args.password), bcrypt.DefaultCost)
			// Validate correct password against legacy bcrypt hash
			check, rehash := CheckPasswordHash(tt.args.password, string(legacyHash))
			if !check {
				t.Errorf("CheckPasswordHash() failed to validate legacy bcrypt password=%v against hash=%v", tt.args.password, string(legacyHash))
			}
			if !rehash {
				t.Errorf("CheckPasswordHash() did not indicate rehashing for legacy bcrypt password=%v", tt.args.password)
			}

			// Validate incorrect passwords against legacy bcrypt hash
			for _, invalid := range tt.args.invalidInputs {
				check, _ := CheckPasswordHash(invalid, string(legacyHash))
				if check {
					t.Errorf("CheckPasswordHash() improperly validated invalid password=%v against legacy hash=%v", invalid, string(legacyHash))
				}
			}
		})
	}
}
