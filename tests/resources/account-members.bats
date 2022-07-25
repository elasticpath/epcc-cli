#!/usr/bin/env bats
@test "account-members empty get collection" {
	run epcc get account-members
	[ $status -eq 0 ]
}



