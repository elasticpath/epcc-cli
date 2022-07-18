#!/usr/bin/env bats
@test "user-authentication-oidc-profile-infos empty get collection" {
	run epcc get user-authentication-oidc-profile-infos
	[ $status -eq 0 ]
}

