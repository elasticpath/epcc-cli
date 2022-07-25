#!/usr/bin/env bats
@test "password-profiles empty get collection" {
	run epcc get password-profiles
	[ $status -eq 0 ]
}

@test "password-profiles delete-all support" {
	run epcc delete-all password-profiles
	[ $status -eq 0 ]
}


