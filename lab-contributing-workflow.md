# Lab 1 - Contributing Workflow

## Writing & Testing Code in Crossplane

### Clone Crossplane Repo

```console
git clone https://github.com/crossplane/crossplane.git
cd crossplane
```

### First Build

```console
make
```

### Bring up Dev Cluster

```console
./cluster/local/kind.sh up
```

### Deploy Local Build of Crossplane

```console
./cluster/local/kind.sh helm-install
```

### Make Local Changes

Make a local change to the Crossplane code, for example adding a new logging
line in the main entry point:

```patch
diff --git a/cmd/crossplane/core/core.go b/cmd/crossplane/core/core.go
index 17d12a34..28594110 100644
--- a/cmd/crossplane/core/core.go
+++ b/cmd/crossplane/core/core.go
@@ -204,5 +204,7 @@ func (c *startCommand) Run(s *runtime.Scheme, log logging.Logger) error { //noli
 		}
 	}

+	log.Info("Contribfest is starting!", "controller options", o)
+
 	return errors.Wrap(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
 }
```

Then run a build again that will include your local changes:

```console
make
```

### Deploy Your Local Changes

```console
./cluster/local/kind.sh update
./cluster/local/kind.sh restart
```

### Make Changes Faster

Make a second local change now, for example updating your logging line with new
content:

```patch
diff --git a/cmd/crossplane/core/core.go b/cmd/crossplane/core/core.go
index 17d12a34..7c804e8d 100644
--- a/cmd/crossplane/core/core.go
+++ b/cmd/crossplane/core/core.go
@@ -204,5 +204,7 @@ func (c *startCommand) Run(s *runtime.Scheme, log logging.Logger) error { //noli
 		}
 	}

+	log.Info("Contribfest is getting faster!", "controller options", o)
+
 	return errors.Wrap(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
 }
```

### Deploy Changes Faster

Instead of having to build and deploy the entire Crossplane image, we can just
run the controllers directly in proc:

```console
kubectl -n crossplane-system scale deploy crossplane --replicas=0
make run
```

### Run Tests

```console
make test
```

Modify one of the unit tests to purposefully fail, and then run the tests again,
so we can see the failure output:

```patch
diff --git a/internal/version/version_test.go b/internal/version/version_test.go
index 772f731c..2c9075bf 100644
--- a/internal/version/version_test.go
+++ b/internal/version/version_test.go
@@ -42,7 +42,7 @@ func TestInRange(t *testing.T) {
 		"ValidInRange": {
 			reason: "Should return true when a valid semantic version is in a valid range.",
 			args: args{
-				version: "v0.13.0",
+				version: "v0.3.0",
 				r:       ">0.12.0",
 			},
 			want: want{
```

```console
make test
```

### Run e2e Integration Tests

```console
make e2e
```

### Getting Changes Ready for Review

```console
make reviewable
```

### Open a PR

Push your changes to a branch on your fork, then open a PR to [upstream
Crossplane](https://github.com/crossplane/crossplane/pulls).  Make sure you fill
out the entire [PR
template](https://github.com/crossplane/crossplane/blob/master/.github/PULL_REQUEST_TEMPLATE.md)
to provide complete context for reviewers.