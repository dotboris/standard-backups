# Delete this file before merging!

- Make use of variant name in restic backend
- Make use of variant name in rsync backend
- Ensure that `forget` (through restic) operates on the current variant and not across all variants
- e2e tests for variants & proto
- e2e tests for restic with variants
- Allow list-backups to operate across all variants
