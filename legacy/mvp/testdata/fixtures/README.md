# Acceptance fixtures

Files below this directory are immutable test inputs.

- Organize each workflow under its own directory.
- Keep source filenames and sheet names when they are part of the contract.
- Store only minimal representative data; do not copy credentials or unnecessary personal data.
- Tests must snapshot every input before execution and assert that it is unchanged afterward.
- Expected outputs belong beside the workflow fixture under an `expected` directory; runtime outputs do not.

