#!/usr/bin/env bats
@test "currencies empty get collection" {
	run epcc get currencies
	[ $status -eq 0 ]
}

