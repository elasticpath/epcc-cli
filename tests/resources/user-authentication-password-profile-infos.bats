#!/usr/bin/env bats
@test "user-authentication-password-profile-infos empty get collection" {
	run epcc get user-authentication-password-profile-infos
	[ $status -eq 0 ]
}

