#!/usr/bin/env bats
@test "accounts empty get collection" {
	run epcc get accounts
	[ $status -eq 0 ]
}

