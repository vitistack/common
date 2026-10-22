package v1alpha1

type MachineProviderType string

const (
	MachineProviderTypeKubevirt MachineProviderType = "kubevirt"
	MachineProviderTypeProxmox  MachineProviderType = "proxmox"
)

func (mpt MachineProviderType) IsValid() bool {
	switch mpt {
	case MachineProviderTypeKubevirt, MachineProviderTypeProxmox:
		return true
	default:
		return false
	}
}

func ValidMachineProviderTypes() []MachineProviderType {
	return []MachineProviderType{
		MachineProviderTypeKubevirt,
		MachineProviderTypeProxmox,
	}
}

func DefaultMachineProviderType() MachineProviderType {
	return MachineProviderTypeKubevirt
}

func MachineProviderTypeValues() []string {
	values := make([]string, len(ValidMachineProviderTypes()))
	for i, mpt := range ValidMachineProviderTypes() {
		values[i] = string(mpt)
	}
	return values
}

func (mpt MachineProviderType) String() string {
	return string(mpt)
}
