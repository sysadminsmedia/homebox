package hasher

import "testing"

func TestHashPassword(t *testing.T) {
	t.Parallel()
	type args struct {
		password string
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		invalidInputs []string
	}{
		{
			name: "letters_and_numbers",
			args: args{
				password: "password123456788",
			},
			invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
		},
		{
			name: "letters_number_and_special",
			args: args{
				password: "!2afj3214pofajip3142j;fa",
			},
			invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
		},
		{
			name: "extra_long_password",
			args: args{
				password: "this_is_a_very_long_password_that_should_be_hashed_properly_and_still_work_with_the_check_function",
			},
			invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
		},
		{
			name: "empty_password",
			args: args{
				password: "",
			},
			invalidInputs: []string{"testPassword", "AnotherBadPassword", "ThisShouldNeverWork", "1234567890"},
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
			if check {
				t.Errorf("CheckPasswordHash() failed to validate password=%v against hash=%v", tt.args.password, got)
			}

			for _, invalid := range tt.invalidInputs {
				check, _ := CheckPasswordHash(invalid, got)
				if check {
					t.Errorf("CheckPasswordHash() improperly validated password=%v against hash=%v", invalid, got)
				}
			}
		})
	}
}
