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

Make a local change to the Crossplane code, for example adding a new logging line in the main entry point:

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

Make a second local change now, for example updating your logging line with new content:

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

Instead of having to build and deploy the entire Crossplane image, we can just run the controllers directly in proc:

```console
kubectl -n crossplane-system scale deploy crossplane --replicas=0
make run
```
