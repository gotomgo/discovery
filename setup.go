package discovery

func InitializeDiscovery() error {
	d := GetOrCreateDefaultDiscovery(nil)

	if err := AddItem(ItemDiscoveryManagementType, d); err != nil {
		return err
	}

	if err := AddItem(ItemResolverType, GetDefaultResolver()); err != nil {
		return err
	}

	return nil
}
