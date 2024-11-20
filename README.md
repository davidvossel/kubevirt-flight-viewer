# kubevirt-flight-viewer


View cluster in-flight operations

Sample output.

```
$ oc get inflightoperations  -w
OPERATION_TYPE   RESOURCE_KIND            RESOURCE_NAME    PROGRESSING
Starting         VirtualMachineInstance   vmi-asdf         True
Starting         VirtualMachineInstance   vmi-sfdg         True
LiveMigration    VirtualMachineInstance   vmi-jkla         True
LiveMigration    VirtualMachineInstance   vmi-vkls         True
Stopping         VirtualMachineInstance   vmi-bvxd         True
Stopping         VirtualMachineInstance   vmi-lsdf         True
Stopping         VirtualMachineInstance   vmi-bfad         True
Stopping         VirtualMachineInstance   vmi-icmn         True
Provisioning     VirtualMachine           vm-jdgr          True
Provisioning     VirtualMachine           vm-lorf          True
DiskHotplug      VirtualMachine           vm-uehj          True
MemoryHotplug    VirtualMachine           vm-qwed          True
```
