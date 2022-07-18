#!/usr/bin/env bats
@test "merchant-realm-mappings empty get collection" {
	run epcc get merchant-realm-mappings
	[ $status -eq 0 ]
}

