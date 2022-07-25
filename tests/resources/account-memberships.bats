#!/usr/bin/env bats
@test "account-memberships empty get collection" {
	run epcc get account-memberships
	[ $status -eq 0 ]
}

@test "account-memberships delete-all support" {
	run epcc delete-all account-memberships
	[ $status -eq 0 ]
}


