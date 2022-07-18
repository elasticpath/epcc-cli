#!/usr/bin/env bats
@test "account-memberships empty get collection" {
	run epcc get account-memberships
	[ $status -eq 0 ]
}

