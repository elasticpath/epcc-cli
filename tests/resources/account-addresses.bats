#!/usr/bin/env bats
@test "account-addresses empty get collection" {
	run epcc get account-addresses
	[ $status -eq 0 ]
}

@test "account-addresses delete-all support" {
	run epcc delete-all account-addresses
	[ $status -eq 0 ]
}


