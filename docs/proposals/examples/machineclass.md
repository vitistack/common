# `MachineClass`

| | |
|---|---|
| **API** | `vitistack.io/v1alpha3` |
| **Scope** | Cluster |
| **Spec written by** | Platform admin |
| **Status written by** | The machine provider matching `spec.controllerName` |
| **Proposal** | [Main proposal](../kubernetes-native-cluster-machine-api.md) §5.1 |

Size plus backend: what a Machine gets and where it runs. The same size on another backend is another class.

## Full specification

```yaml
apiVersion: vitistack.io/v1alpha3
kind: MachineClass
metadata:
  name: gpu-xlarge
spec:
  # Required. Controller that implements Machines of this class. Immutable.
  controllerName: kubevirt.vitistack.io/machine-controller
  # Required. Provider registration (§3.1): which backend, e.g. one KubeVirt cluster.
  machineProviderRef:
    name: kubevirt-prod-1
  # Required. Size of each Machine.
  resources:
    cpu: 16
    memory: 64Gi
    # Optional. Root disk size.
    disk: 100Gi
  # Optional. Provider-specific parameters.
  parametersRef:
    group: kubevirt.vitistack.io
    kind: KubevirtMachineParameters
    name: gpu
status:
  observedGeneration: 1
  conditions:
    # Accepted=False reasons: ProviderMismatch, ProviderNotAvailable, ParametersNotFound.
    - { type: Accepted, status: "True", reason: Accepted }
```

## Minimal

```yaml
apiVersion: vitistack.io/v1alpha3
kind: MachineClass
metadata:
  name: medium
spec:
  controllerName: kubevirt.vitistack.io/machine-controller
  machineProviderRef:
    name: kubevirt-prod-1
  resources:
    cpu: 4
    memory: 16Gi
```
