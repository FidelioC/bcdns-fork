# Testing Guide Verifying Network

This readme includes unit test for the verifying network extension. The tests are designed to verify the functionality of the verifying network.

## Running Tests

To specifically run the verifying network tests:

- make sure that you're in the right folder path, /dns_client_golang/verifying_network:

```bash
go test . -v
```

### Running Tests for a Specific Component

To run tests for a specific component, you can use the following command:

- make sure that you're in the right folder path, /dns_client_golang/verifying_network

```bash
go test -v -run ^Test_Function_Name$
```
