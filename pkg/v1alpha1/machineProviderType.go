package v1alpha1

type MachineProviderType string

const (
	MachineProviderTypeTalos MachineProviderType = "kubevirt"
)

func (mpt MachineProviderType) IsValid() bool {
	switch mpt {
	case MachineProviderTypeTalos:
		return true
	default:
		return false
	}
}

func ValidMachineProviderTypes() []MachineProviderType {
	return []MachineProviderType{
		MachineProviderTypeTalos,
	}
}

func DefaultMachineProviderType() MachineProviderType {
	return MachineProviderTypeTalos
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
