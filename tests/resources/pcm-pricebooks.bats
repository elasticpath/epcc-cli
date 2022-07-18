#!/usr/bin/env bats
@test "pcm-pricebooks empty get collection" {
	run epcc get pcm-pricebooks
	[ $status -eq 0 ]
}

