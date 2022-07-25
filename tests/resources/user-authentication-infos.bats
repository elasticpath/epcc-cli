#!/usr/bin/env bats
@test "user-authentication-infos empty get collection" {
	run epcc get user-authentication-infos
	[ $status -eq 0 ]
}

@test "user-authentication-infos delete-all support" {
	run epcc delete-all user-authentication-infos
	[ $status -eq 0 ]
}


