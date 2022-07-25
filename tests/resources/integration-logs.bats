#!/usr/bin/env bats
@test "integration-logs empty get collection" {
	run epcc get integration-logs
	[ $status -eq 0 ]
}



