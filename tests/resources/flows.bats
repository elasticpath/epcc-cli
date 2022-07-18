#!/usr/bin/env bats
@test "flows empty get collection" {
	run epcc get flows
	[ $status -eq 0 ]
}

