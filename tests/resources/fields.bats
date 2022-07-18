#!/usr/bin/env bats
@test "fields empty get collection" {
	run epcc get fields
	[ $status -eq 0 ]
}

