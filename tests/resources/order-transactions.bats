#!/usr/bin/env bats
@test "order-transactions empty get collection" {
	run epcc get order-transactions
	[ $status -eq 0 ]
}

