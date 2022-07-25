#!/usr/bin/env bats
@test "authentication-realms empty get collection" {
	run epcc get authentication-realms
	[ $status -eq 0 ]
}



