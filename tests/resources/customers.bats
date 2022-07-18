#!/usr/bin/env bats
@test "customers empty get collection" {
	run epcc get customers
	[ $status -eq 0 ]
}

