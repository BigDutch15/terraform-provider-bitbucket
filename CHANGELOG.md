## Unreleased

### Notes

* The provider code is now the same as [DrFaust92/bitbucket v2.57.0](https://github.com/DrFaust92/terraform-provider-bitbucket/releases/tag/v2.57.0). This fork's own implementations of `require_all_comments_resolved`, `require_review_group_approvals_to_merge`, `pipelines_source_branches` and workspace UUID resolution for the group resources are replaced by the upstream versions ([#7](https://github.com/BigDutch15/terraform-provider-bitbucket/issues/7))

### Features

* `enforced_signed_commits` on `bitbucket_repository`, `bitbucket_project` and their data sources; `is_personal`, `is_privacy_enforced`, `forking_mode` and `enforced_signed_commits` on the `bitbucket_workspace` data source; `fingerprint`, `expires_on` and `last_used` on `bitbucket_ssh_key`; `added_on`, `expires_on` and `last_used` on `bitbucket_gpg_key` (from upstream)
* **New resource:** `bitbucket_pipeline_runner`, a self-hosted Pipelines runner for a workspace or a repository (from upstream)
* Provider: new `api_url`, `timeout` and `max_retries` arguments (`BITBUCKET_API_URL`, `BITBUCKET_TIMEOUT`, `BITBUCKET_MAX_RETRIES`) (from upstream)

## 2.56.0 (November 24, 2026)

### Notes

* Includes the upstream changes released as [DrFaust92/bitbucket v2.55.2](https://github.com/DrFaust92/terraform-provider-bitbucket/releases/tag/v2.55.2), merged in [#24](https://github.com/BigDutch15/terraform-provider-bitbucket/pull/24). Version numbers no longer track upstream: upstream's v2.55.1 and v2.55.2 are different releases from this provider's v2.55.1.

### Features

* `bitbucket_branch_restriction`: support the `require_all_comments_resolved` kind ([#5](https://github.com/BigDutch15/terraform-provider-bitbucket/issues/5))
* `bitbucket_branch_restriction`: support the `require_review_group_approvals_to_merge` kind and the `pipelines_source_branches` argument ([#8](https://github.com/BigDutch15/terraform-provider-bitbucket/issues/8))
  * Read now uses the provider's raw HTTP client, like create and update, because the generated client has no `pipelines_source_branches` field. Read errors now use the `API Error: <status> <endpoint> <message>` format.

### Bug fixes

* `bitbucket_group`, `bitbucket_group_membership`, `data.bitbucket_group`, `data.bitbucket_groups`, `data.bitbucket_group_members`: `workspace` now accepts the workspace UUID as well as the slug. A UUID is resolved to the slug before calling the 1.0 groups API, which previously rejected it with HTTP 403 or 405. The docs now use `data.bitbucket_workspace.<name>.slug` ([#14](https://github.com/BigDutch15/terraform-provider-bitbucket/issues/14))
* `bitbucket_branch_restriction`: `value = 0` is now sent to Bitbucket on create and update instead of being dropped, so restrictions such as `require_commits_behind` with `value = 0` no longer fail with HTTP 400 ([#19](https://github.com/BigDutch15/terraform-provider-bitbucket/issues/19))
  * Create and update errors now use the provider's common `API Error: <status> <endpoint> <message>` format instead of `<status>: <message>`.
* Provider: each Bitbucket API request now times out after 2 minutes, so an unresponsive endpoint can no longer hang a `plan` or `apply` indefinitely. Requests that are retried get 2 minutes per attempt (from upstream [DrFaust92/terraform-provider-bitbucket#273](https://github.com/DrFaust92/terraform-provider-bitbucket/pull/273))

## 2.55.1 (November 16, 2026)

### Notes

* The provider is now published as `BigDutch15/bitbucket` from [BigDutch15/terraform-provider-bitbucket](https://github.com/BigDutch15/terraform-provider-bitbucket). The code is the same as `DrFaust92/bitbucket` v2.55.0. To switch, change `source` to `BigDutch15/bitbucket` in `required_providers`, then run `terraform state replace-provider DrFaust92/bitbucket BigDutch15/bitbucket` and `terraform init`.

## 2.55.0 (DrFaust92/bitbucket)

Releases 1.3.0 through 2.55.0 were published by [DrFaust92/terraform-provider-bitbucket](https://github.com/DrFaust92/terraform-provider-bitbucket/releases); see its release notes for those changes.

### Bug fixes

* `bitbucket_branch_restriction`: `users` and `groups` previously failed to populate in state at all, because the Bitbucket API's GDPR changes removed usernames from account responses while the provider still expected them. Now: `users` round-trips via UUID, and `groups.owner` round-trips via workspace slug or UUID (whichever form is already in your config/state). **Behavior change:** a config that lists a username in `users` previously showed no diff (because `Read` silently failed to populate state); it will now show a diff on every `plan`/`apply` until you switch that entry to the UUID shown in the new warning, or add `lifecycle { ignore_changes = [users] }`.

## 1.3.0 (March 15, 2021)

This release contains the changes to upstream repo that were never released.

### Features

* add `bitbucket_deployment` and `bitbucket_deployment_variable` resources [#60](https://github.com/hashicorp/terraform-provider-bitbucket/pull/60)
* add `require_default_reviewer_approvals_to_merge` branch restriction value [#52](https://github.com/hashicorp/terraform-provider-bitbucket/pull/52)

### Bug fixes

* fix issue with omitempty [#49](https://github.com/hashicorp/terraform-provider-bitbucket/pull/49)

### Documentation

* fix ducmentation typo [#54](https://github.com/hashicorp/terraform-provider-bitbucket/pull/54), [#61](https://github.com/hashicorp/terraform-provider-bitbucket/pull/61) and [#65](https://github.com/hashicorp/terraform-provider-bitbucket/pull/65)

## 1.2.0 (January 23, 2020)
* add `bitbucket_project` to create a new project via the API
* add `bitbucket_repository` turn on/off pipelines
* add `bitbucket_repository_variable` to add variables via terraform to your pipelines builds
* add `bitbucket_user` to find a user and use for default reviewers.

## 1.1.0 (June 19, 2019)

### Features

* add `skip_cert_verification` for hooks [#19](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/19)

### Bug fixes

* handle missing hooks [#24](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/24)
* fix default reviewer pagination bug [#28](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/28)

### Dev updates

* add `website` and `website-test` targets [#16](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/16)
* add `website-test` target to Travis [#17](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/17)
* upgrade to go 1.11 [#25](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/25)
* switch to go modules [#27](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/27)
* upgrade to `hashicorp/terraform` v0.12.2 [#34](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/34)

### Documentation

* add note about v1 APIs [#21](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/21)

## 1.0.0 (December 08, 2017)

* resource/bitbucket_repository: Add the ability to define a seperate slug for a repository ([#5](https://github.com/terraform-providers/terraform-provider-bitbucket/issues/5))

## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
