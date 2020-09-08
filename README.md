# romver-resource
[![GitHub release](https://img.shields.io/github/release/cappyzawa/romver-resource.svg)](https://github.com/cappyzawa/romver-resource/releases)
[![GitHub](https://img.shields.io/github/license/cappyzawa/romver-resource.svg)](./LICENSE)

A resource for managing a version number. Persists the version number in one of several backing stores.

This resource can manage major version(`X`) only.

When version is bumped by this resouce, new version is `X+1`.

There is no concept of `major`, `minor` and `patch` like semanic versioning.

## Source Configuration

* `driver`: *Required.* The driver to use for tracking the
  version. Determines where the version is stored. (`git` only, yet)

* `initial_version`: *Optional.* The version number to use when
bootstrapping, i.e. when there is not a version number present in the source.

### `git` Driver

The `git` driver works by modifying a file in a repository with every bump. The
`git` driver has the advantage of being able to do atomic updates.

* `uri`: *Required.* The repository URL.

* `branch`: *Required.* The branch the file lives on.

* `file`: *Required.* The name of the file in the repository.

* `private_key`: *Optional.* The SSH private key to use when pulling from/pushing to to the repository.

* `username`: *Optional.* Username for HTTP(S) auth when pulling/pushing.
   This is needed when only HTTP/HTTPS protocol for git is available (which does not support private key auth)
   and auth is required.

* `password`: *Optional.* Password for HTTP(S) auth when pulling/pushing.

* `git_user`: *Optional.* The git identity to use when pushing to the
  repository support RFC 5322 address of the form "Gogh Fir \<gf@example.com\>" or "foo@example.com".

* `depth`: *Optional.* If a positive integer is given, shallow clone the repository using the --depth option.

* `commit_message`: *Optional.* If specified overides the default commit message with the one provided. The user can use %version% and %file% to get them replaced automatically with the correct values.

### Example

With the following resource configuration:

``` yaml
resource_types:
- name: romver
  type: registry-image
  source:
    repository: ghcr.io/cappyzawa/romver-resource
resources:
- name: version
  type: romver
  source:
    driver: git
    uri: git@github.com:concourse/concourse.git
    branch: version
    file: version
    private_key: ((concourse-repo-private-key))
```

Bumping with a `get` and then a `put`:

``` yaml
plan:
- get: version
  params: {bump: true}
- task: a-thing-that-needs-a-version
- put: version
  params: {file: version/version}
```

Or, bumping with an atomic `put`:

``` yaml
plan:
- put: version
  params: {bump: true}
- task: a-thing-that-needs-a-version
```

## Behavior

### `check`: Report the current version number.

Detects new versions by reading the file from the specified source. If the file is empty, it returns the `initial_version`. If the file is not empty, it returns the version specified in the file if it is equal to or greater than current version, otherwise it returns no versions.

### `in`: Provide the version as a file, optionally bumping it.

Provides the version number to the build as a `version` file in the destination.

#### Parameters

* `bump`: *Optional.* `true` or `false`

### `out`: Set the version or bump the current one.

Given a file, use its contents to update the version. Or, given a bump
strategy, bump whatever the current version is. If there is no current version,
the bump will be based on `initial_version`.

The `file` parameter should be used if you have a particular version that you
want to force the current version to be. This can be used in combination with
`in`, but it's probably better to use the `bump` params as they'll
perform an atomic in-place bump if possible with the driver.

#### Parameters

One of the following must be specified:

* `file`: *Optional.* Path to a file containing the version number to set.

* `bump`: *Optional.* `true` or `false`

When `bump` used, the version bump will be applied atomically,
if the driver supports it. That is, if we pull down version `N`. 

### Running the tests
```
$ cd romver-resource
$ fly -t <target> -c ci/tasks/test.yml -i romver-resource=. 
```

#### Integration tests

The integration requires two AWS S3 buckets, one without versioning and another
with. The `docker build` step requires setting `--build-args` so the
integration will run.

You will need:
* Github uri and branch
* Github username and password

Run the tests with the following command, replacing each `build-arg` value with your own values:

```sh
docker build . -t semver-resource \
    --build-arg ROMVER_TESTING_GITHUB_URI='https://github.com/your/repo' \
    --build-arg ROMVER_TESTING_GITHUB_BRANCH='branch' \
    --build-arg ROMVER_TESTING_GITHUB_USERNAME='github-username' \
    --build-arg ROMVER_TESTING_GITHUB_PASSWORD='github-password'
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
