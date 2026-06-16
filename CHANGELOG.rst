================================================
The Terraform Provider for Ansible Release Notes
================================================

.. contents:: Topics

v1.5.0
======

Major Changes
-------------

- Add ansible_vault data sources and ephemeral resources (https://github.com/ansible/terraform-provider-ansible/pull/156).

Minor Changes
-------------

- Bootstrap unit and acceptance test infrastructure (https://github.com/ansible/terraform-provider-ansible/pull/158).
- Bump CI dependencies (https://github.com/ansible/terraform-provider-ansible/pull/162).
- Fix documentation for ansible_playbook_run (https://github.com/ansible/terraform-provider-ansible/pull/163).
- Preallocate slices with known capacity for improved performance (https://github.com/ansible/terraform-provider-ansible/pull/159).
- Update Go version to 1.26 and bump dependencies (https://github.com/ansible/terraform-provider-ansible/pull/157).

Bugfixes
--------

- fix(action) - Skip file validation for unknown/null values in playbook_run (https://github.com/ansible/terraform-provider-ansible/pull/155).

v1.4.0
======

Major Changes
-------------

- Add Terraform Action support to playbooks.
- Add support for complex types in ansible_host and ansible_group variables.

Bugfixes
--------

- Fix 'make test' to run successfully.

v1.3.0
======

Minor Changes
-------------

- resource/ansible_playbook - Provider should failed with proper message when ansible is not installed (https://github.com/ansible/terraform-provider-ansible/issues/35).

Bugfixes
--------

- ensure extra vars are quoted (https://github.com/ansible/terraform-provider-ansible/pull/57).

v1.2.0
======

Release Summary
---------------

The terraform-provider-ansible v1.2.0 includes minor bugfixes and improvements.

Minor Changes
-------------

- Update dependencies (google.golang.org/grpc and golang.org/x/net) to resolve security alerts https://github.com/ansible/terraform-provider-ansible/security/dependabot (https://github.com/ansible/terraform-provider-ansible/pull/72).
- Updates the provider to use Go 1.21 (https://github.com/ansible/terraform-provider-ansible/pull/89)
- Updates the provider to use SDKv2 (https://github.com/ansible/terraform-provider-ansible/issues/39).

Bugfixes
--------

- provider/resource_playbook - Fix race condition between multiple ansible_playbook resources (https://github.com/ansible/terraform-provider-ansible/issues/38).
