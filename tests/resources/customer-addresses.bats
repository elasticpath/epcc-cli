#!/usr/bin/env bats
@test "customer-addresses empty get collection" {
	run epcc get customer-addresses
	[ $status -eq 0 ]
}

