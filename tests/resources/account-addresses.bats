#!/usr/bin/env bats
@test "account-addresses empty get collection" {
	run epcc get account-addresses
	[ $status -eq 0 ]
}

