#!/usr/bin/env bats
@test "customer-addresses empty get collection" {
	run epcc get customer-addresses
	[ $status -eq 0 ]
}

@test "customer-addresses delete-all support" {
	run epcc delete-all customer-addresses
	[ $status -eq 0 ]
}


