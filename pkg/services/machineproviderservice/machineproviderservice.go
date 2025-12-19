package machineproviderservice

import (
	"context"
	"fmt"

	vitistackv1alpha1 "github.com/vitistack/common/pkg/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineProviderService provides methods to fetch and manage MachineProvider resources
type MachineProviderService struct {
	client client.Client
}

// NewMachineProviderService creates a new MachineProviderService
func NewMachineProviderService(c client.Client) *MachineProviderService {
	return &MachineProviderService{
		client: c,
	}
}

// FetchAllMachineProviders returns all MachineProvider resources (cluster-scoped)
func (s *MachineProviderService) FetchAllMachineProviders(ctx context.Context) ([]vitistackv1alpha1.MachineProvider, error) {
	machineProviderList := &vitistackv1alpha1.MachineProviderList{}

	if err := s.client.List(ctx, machineProviderList); err != nil {
		return nil, fmt.Errorf("failed to list MachineProviders: %w", err)
	}

	return machineProviderList.Items, nil
}

// FetchMachineProviderByName returns a specific MachineProvider by name (cluster-scoped)
func (s *MachineProviderService) FetchMachineProviderByName(ctx context.Context, name string) (*vitistackv1alpha1.MachineProvider, error) {
	machineProvider := &vitistackv1alpha1.MachineProvider{}

	key := client.ObjectKey{
		Name: name,
	}

	if err := s.client.Get(ctx, key, machineProvider); err != nil {
		return nil, fmt.Errorf("failed to get MachineProvider %s: %w", name, err)
	}

	return machineProvider, nil
}

// CreateMachineProvider creates a new MachineProvider resource (cluster-scoped)
func (s *MachineProviderService) CreateMachineProvider(ctx context.Context, machineProvider *vitistackv1alpha1.MachineProvider) error {
	if err := s.client.Create(ctx, machineProvider); err != nil {
		return fmt.Errorf("failed to create MachineProvider %s: %w", machineProvider.Name, err)
	}
	return nil
}
