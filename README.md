# Kubectl Confirm Plugin

Kubectl Confirm is a plugin for Kubectl that displays information and asked for confirmation before executing a command.

The following information is displayed:
* Configuration: Context name, Cluster, User, and Namespace
* Dry Run Output (if the executed command supports the `--dry-run` flag)
* Diff Output (if the executed command supports the `--dry-run` and `--output` flags)

## Example Output
```
$ kubectl confirm apply -f ~/changed.yaml
========== Config ===========
Context:    kind-kind
Cluster:    kind-kind
User:       kind-kind
Namespace:  default

========== Dry Run ==========
deployment.apps/foo configured (server dry run)

========== Diff =============
diff -u -N /tmp/LIVE-2275701238/apps.v1.Deployment.default.foo /tmp/MERGED-1055657464/apps.v1.Deployment.default.foo
--- /tmp/LIVE-2275701238/apps.v1.Deployment.default.foo	2022-07-28 10:27:25.690604172 -0400
+++ /tmp/MERGED-1055657464/apps.v1.Deployment.default.foo	2022-07-28 10:27:25.694604038 -0400
@@ -6,7 +6,7 @@
     kubectl.kubernetes.io/last-applied-configuration: |
       {"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{},"name":"foo","namespace":"default"},"spec":{"replicas":1,"revisionHistoryLimit":10,"selector":{"matchLabels":{"app":"foo"}},"template":{"metadata":{"labels":{"app":"foo"}},"spec":{"containers":[{"command":["sh","-c","echo Container bar is running! \u0026\u0026 sleep 9999999"],"image":"busybox:1.30","name":"bar"}]}}}}
   creationTimestamp: "2022-07-28T11:43:58Z"
-  generation: 1
+  generation: 2
   managedFields:
   - apiVersion: apps/v1
     fieldsType: FieldsV1
@@ -90,7 +90,7 @@
 spec:
   progressDeadlineSeconds: 600
   replicas: 1
-  revisionHistoryLimit: 10
+  revisionHistoryLimit: 11
   selector:
     matchLabels:
       app: foo

========== Confirm ==========
The following command will be executed:
kubectl apply -f /home/bpursley/changed.yaml

Enter 'yes' to continue: 
```

If you enter `yes` (exactly) then it will proceed:
```
Enter 'yes' to continue: yes

deployment.apps/foo configured
```

If you enter anything other than `yes` (exactly), then it will stop:
```
Enter 'yes' to continue: no

Command aborted.
```

## Known Limitations

* Command line completion does not work

