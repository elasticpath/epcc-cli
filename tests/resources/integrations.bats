#!/usr/bin/env bats
@test "integrations empty get collection" {
	run epcc get integrations
	[ $status -eq 0 ]
}

