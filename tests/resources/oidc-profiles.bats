#!/usr/bin/env bats
@test "oidc-profiles empty get collection" {
	run epcc get oidc-profiles
	[ $status -eq 0 ]
}

