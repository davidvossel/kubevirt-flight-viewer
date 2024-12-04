# kubevirt-flight-viewer


View cluster in-flight operations

Below is sample output of inflight operations occurring on a cluster. This
example demonstrates how the controller can provide information about both
VMs and other objects within the cluster which are in a transition state.

In this case, we can see that the oadp operator is updating, and that there
are multiple VMs in various transitory states.

```
$ oc get inflightoperations -A -w

NAMESPACE       NAME        OPERATION_TYPE   RESOURCE_KIND           RESOURCE_NAME          AGE   MESSAGE
openshift-adp   ifo-k68vr   Installing       ClusterServiceVersion   oadp-operator.v1.4.1   20s    
openshift-adp   ifo-x6hcv   Replacing        ClusterServiceVersion   oadp-operator.v1.4.0   2s    phase [Replacing]: being replaced by csv: oadp-operator.v1.4.1
default         ifo-adkef   Starting         VirtualMachineInstance  vmi-1                  10s
default         ifo-jd93s   Starting         VirtualMachineInstance  vmi-2                  7s
default         ifo-d03kd   LiveMigrating    VirtualMachineInstance  vmi-3                  16s
default         ifo-ad3xc   LiveMigrating    VirtualMachineInstance  vmi-4                  9s
default         ifo-93mda   Stopping         VirtualMachineInstance  vmi-5                  23s
default         ifo-ldnad   DiskHotplug      VirtualMachineInstance  vmi-6                  3s
default         ifo-ldnad   MemoryHotPlug    VirtualMachineInstance  vmi-7                  10s
default         ifo-sdadf   LiveMigrating    VirtualMachineInstance  vmi-7                  10s
```

Below is an example of sample yaml for an inflight operation.

```
apiVersion: kubevirtflightviewer.kubevirt.io/v1alpha1
kind: InFlightOperation
metadata:
  creationTimestamp: "2024-12-04T20:01:57Z"
  generateName: ifo-
  generation: 1
  name: ifo-dn94j
  namespace: openshift-cnv
  ownerReferences:
  - apiVersion: operators.coreos.com/v1alpha1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceVersion
    name: kubevirt-hyperconverged-operator.4.16.5-78
    uid: d00a1459-f741-46e8-b278-0aff9f5ac015
  resourceVersion: "236164"
  uid: 676dcb03-b78f-491d-bf61-2b29e13359d9
status:
  operationState:
    message: 'phase [Succeeded]: install strategy completed with no errors'
    reason: Installing
    transitionState: Progressing
  operationType: Installing
  resourceReference:
    apiVersion: operators.coreos.com/v1alpha1, Kind=ClusterServiceVersion
    kind: ClusterServiceVersion
    name: kubevirt-hyperconverged-operator.4.16.5-78
    namespace: openshift-cnv
    uid: d00a1459-f741-46e8-b278-0aff9f5ac015
```
