#!/usr/bin/env bats

#setup_file() {
#  #run epcc reset-store .+
#  #[ $status -eq 0 ]
#}
setup() {
  load 'test_helper/bats-support/load' # this is required by bats-assert!
  load 'test_helper/bats-assert/load'

}

@test "account-addresses crud smoke test" {

  # Fixture Setup
  run epcc delete-all accounts
  [ $status -eq 0 ]
  run epcc delete-all account-addresses
  [ $status -eq 0 ]
  run epcc create account name account-addresses-test legal_name my_legal_name
  [ $status -eq 0 ]

  # Execute SUT
  [ $status -eq 0 ]
	run epcc get account-addresses name=account-addresses-test
	[ $status -eq 0 ]
	run epcc create account-addresses name=account-addresses-test name test legal_name my_legal_name
	[ $status -eq 0 ]
	run epcc update account-addresses name=account-addresses-test name=test name test2
	[ $status -eq 0 ]
	run epcc get account-addresses name=account-addresses-test
	[ $status -eq 0 ]
  assert_output --partial '"name": "test2",'
  run epcc get account-addresses name=test2
  assert_output --partial '"legal_name": "my_legal_name",'

  # Verification



}


@test "accounts crud smoke test" {
  run epcc delete-all accounts
  [ $status -eq 0 ]
	run epcc get accounts
	[ $status -eq 0 ]
	run epcc create account name test legal_name my_legal_name
	[ $status -eq 0 ]
	run epcc update account name=test name test2
	[ $status -eq 0 ]
	run epcc get accounts
	[ $status -eq 0 ]
  assert_output --partial '"name": "test2",'
  run epcc get account name=test2
  assert_output --partial '"legal_name": "my_legal_name",'
}


