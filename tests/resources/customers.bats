#!/usr/bin/env bats
@test "customers empty get collection" {
	run epcc get customers
	[ $status -eq 0 ]
}

@test "customers delete-all support" {
	run epcc delete-all customers
	[ $status -eq 0 ]
}


