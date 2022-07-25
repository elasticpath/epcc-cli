#!/usr/bin/env bats
@test "oidc-profiles empty get collection" {
	run epcc get oidc-profiles
	[ $status -eq 0 ]
}

@test "oidc-profiles delete-all support" {
	run epcc delete-all oidc-profiles
	[ $status -eq 0 ]
}


