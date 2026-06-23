package models

import "testing"

func TestRoleCanDeleteTeam(t *testing.T) {
	cases := []struct {
		role Role
		want bool
	}{
		{role: RoleOwner, want: true},
		{role: RoleAdmin, want: false},
		{role: RoleMember, want: false},
	}

	for _, tc := range cases {
		if got := tc.role.CanDeleteTeam(); got != tc.want {
			t.Fatalf("CanDeleteTeam(%q) = %v, want %v", tc.role, got, tc.want)
		}
	}
}
