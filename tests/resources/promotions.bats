#!/usr/bin/env bats
@test "promotions empty get collection" {
	run epcc get promotions
	[ $status -eq 0 ]
}

