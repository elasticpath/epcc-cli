#!/usr/bin/env bats
@test "user-authentication-infos empty get collection" {
	run epcc get user-authentication-infos
	[ $status -eq 0 ]
}

