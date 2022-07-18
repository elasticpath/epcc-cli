#!/usr/bin/env bats
@test "password-profiles empty get collection" {
	run epcc get password-profiles
	[ $status -eq 0 ]
}

